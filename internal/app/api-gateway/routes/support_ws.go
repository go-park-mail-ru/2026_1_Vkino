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

	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
	wspkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/ws"
)

type supportWSContextKey string

const supportWSClientIDContextKey supportWSContextKey = "support_ws_client_id"

func newSupportTicketSubscribeHandler(userClient UserClient) http.HandlerFunc {
	var nextClientID atomic.Int64

	hubs := wspkg.NewGroup()
	streamCancels := &sync.Map{}

	return func(w http.ResponseWriter, r *http.Request) {
		ticketID, ok := parsePathID(w, r, "invalid ticket id")
		if !ok {
			return
		}

		authorization := supportWSAuthorization(r)
		if _, err := authctx.ParseBearerToken(authorization); err != nil {
			httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		streamCtx := authctx.AppendOutgoing(r.Context(), authorization)
		streamCtx, cancelStream := context.WithCancel(streamCtx)

		stream, err := userClient.SubscribeTicket(streamCtx, &supportv1.SubscribeTicketRequest{
			TicketId: ticketID,
		})
		if err != nil {
			cancelStream()
			writeGRPCError(w, err)

			return
		}

		hub := hubs.Hub(strconv.FormatInt(ticketID, 10))
		clientID := nextClientID.Add(1)

		ctx := context.WithValue(r.Context(), supportWSClientIDContextKey, clientID)
		*r = *r.WithContext(ctx)

		connected := false
		defer func() {
			if !connected {
				cancelStream()
			}
		}()

		handler := wspkg.ServeWS(wspkg.HTTPUpgrader{}, hub, wspkg.ServeOptions{
			SendBuffer: wspkg.DefaultSendBuffer,
			ClientID: func(r *http.Request) (int64, error) {
				clientID, ok := r.Context().Value(supportWSClientIDContextKey).(int64)
				if !ok || clientID == 0 {
					return 0, wspkg.ErrClientIDRequired
				}

				return clientID, nil
			},
			OnConnect: func(ctx context.Context, client *wspkg.Client) error {
				connected = true
				streamCancels.Store(client.ID(), cancelStream)

				go func() {
					requestLogger := logger.FromContext(ctx).
						WithField("ticket_id", ticketID).
						WithField("ws_client_id", client.ID())

					for {
						event, recvErr := stream.Recv()
						if recvErr != nil {
							if !errors.Is(recvErr, io.EOF) && !errors.Is(recvErr, context.Canceled) {
								requestLogger.WithField("error", recvErr).Error("support websocket stream closed with error")
							}

							_ = client.Close()

							return
						}

						payload, marshalErr := json.Marshal(event)
						if marshalErr != nil {
							requestLogger.WithField("error", marshalErr).Error("failed to marshal support websocket event")
							_ = client.Close()

							return
						}

						if sendErr := client.Send(payload); sendErr != nil {
							requestLogger.WithField("error", sendErr).Error("failed to send support websocket event")
							_ = client.Close()

							return
						}
					}
				}()

				return nil
			},
			OnClose: func(ctx context.Context, client *wspkg.Client, err error) {
				if cancel, ok := streamCancels.LoadAndDelete(client.ID()); ok {
					cancel.(context.CancelFunc)()
				}

				if err != nil {
					logger.FromContext(ctx).
						WithField("ticket_id", ticketID).
						WithField("ws_client_id", client.ID()).
						WithField("error", err).
						Error("support websocket closed with error")
				}
			},
		})

		handler(w, r)
	}
}

func supportWSAuthorization(r *http.Request) string {
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
