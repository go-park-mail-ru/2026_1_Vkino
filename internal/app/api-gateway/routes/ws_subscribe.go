package routes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
	wspkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/ws"
)

type subscribeStream[T any] interface {
	Recv() (*T, error)
}

type wsSubscribeConfig[T any] struct {
	hubKey        int64
	resourceName  string
	resourceField string
	clientIDKey   any
	request       *http.Request
	stream        subscribeStream[T]
	cancelStream  context.CancelFunc
	streamCancels *sync.Map
	nextClientID  *atomic.Int64
	hubs          *wspkg.Group
	streamErrLog  string
	marshalErrLog string
	sendErrLog    string
	closeErrLog   string
}

func serveSubscription[T any](w http.ResponseWriter, cfg wsSubscribeConfig[T]) {
	hub := cfg.hubs.Hub(strconv.FormatInt(cfg.hubKey, 10))
	clientID := cfg.nextClientID.Add(1)

	ctx := context.WithValue(cfg.request.Context(), cfg.clientIDKey, clientID)
	*cfg.request = *cfg.request.WithContext(ctx)

	connected := false

	defer func() {
		if !connected {
			cfg.cancelStream()
		}
	}()

	handler := wspkg.ServeWS(wspkg.HTTPUpgrader{}, hub, wspkg.ServeOptions{
		SendBuffer: wspkg.DefaultSendBuffer,
		ClientID: func(r *http.Request) (int64, error) {
			clientID, ok := r.Context().Value(cfg.clientIDKey).(int64)
			if !ok || clientID == 0 {
				return 0, wspkg.ErrClientIDRequired
			}

			return clientID, nil
		},
		OnConnect: func(ctx context.Context, client *wspkg.Client) error {
			connected = true

			cfg.streamCancels.Store(client.ID(), cfg.cancelStream)

			go streamSubscriptionEvents(ctx, client, cfg)

			return nil
		},
		OnClose: func(ctx context.Context, client *wspkg.Client, err error) {
			cancelSubscriptionClient(cfg.streamCancels, client.ID())

			if err != nil {
				logger.FromContext(ctx).
					WithField(cfg.resourceField, cfg.hubKey).
					WithField("ws_client_id", client.ID()).
					WithField("error", err).
					Error(cfg.closeErrLog)
			}
		},
	})

	handler(w, cfg.request)
}

func streamSubscriptionEvents[T any](ctx context.Context, client *wspkg.Client, cfg wsSubscribeConfig[T]) {
	requestLogger := logger.FromContext(ctx).
		WithField(cfg.resourceField, cfg.hubKey).
		WithField("ws_client_id", client.ID())

	for {
		event, recvErr := cfg.stream.Recv()
		if recvErr != nil {
			if !errors.Is(recvErr, io.EOF) && !errors.Is(recvErr, context.Canceled) {
				requestLogger.WithField("error", recvErr).Error(cfg.streamErrLog)
			}

			_ = client.Close()

			return
		}

		payload, marshalErr := json.Marshal(event)
		if marshalErr != nil {
			requestLogger.WithField("error", marshalErr).Error(cfg.marshalErrLog)

			_ = client.Close()

			return
		}

		if sendErr := client.Send(payload); sendErr != nil {
			requestLogger.WithField("error", sendErr).Error(cfg.sendErrLog)

			_ = client.Close()

			return
		}
	}
}

func cancelSubscriptionClient(streamCancels *sync.Map, clientID int64) {
	if cancelValue, ok := streamCancels.LoadAndDelete(clientID); ok {
		cancelFn, castOK := cancelValue.(context.CancelFunc)
		if castOK {
			cancelFn()
		}
	}
}

func subscribeAuthorization(r *http.Request) string {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization != "" {
		return authorization
	}

	accessToken := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if accessToken == "" {
		return ""
	}

	return "Bearer " + accessToken
}

func authorizeSubscription(w http.ResponseWriter, r *http.Request) (string, bool) {
	authorization := subscribeAuthorization(r)
	if _, err := authctx.ParseBearerToken(authorization); err != nil {
		httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

		return "", false
	}

	return authorization, true
}
