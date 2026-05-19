//nolint:gocognit // WS subscription flow is intentionally explicit.
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

	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
	wspkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/ws"
)

type partyWSContextKey string

const partyWSClientIDContextKey partyWSContextKey = "party_ws_client_id"

//nolint:gocyclo,cyclop // Subscription wiring intentionally keeps the control flow explicit.
func newPartyRoomSubscribeHandler(partyClient PartyClient) http.HandlerFunc {
	var nextClientID atomic.Int64

	hubs := wspkg.NewGroup()
	streamCancels := &sync.Map{}

	return func(w http.ResponseWriter, r *http.Request) {
		roomID, ok := parseRoomPathID(w, r, "invalid room id")
		if !ok {
			return
		}

		authorization := wsAuthorization(r)
		if _, err := authctx.ParseBearerToken(authorization); err != nil {
			httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		streamCtx := authctx.AppendOutgoing(r.Context(), authorization)
		streamCtx, cancelStream := context.WithCancel(streamCtx)

		stream, err := partyClient.SubscribeRoom(streamCtx, &partyv1.SubscribeRoomRequest{
			RoomId: roomID,
		})
		if err != nil {
			cancelStream()
			writeGRPCError(w, err)

			return
		}

		hub := hubs.Hub(strconv.FormatInt(roomID, 10))
		clientID := nextClientID.Add(1)

		ctx := context.WithValue(r.Context(), partyWSClientIDContextKey, clientID)
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
				clientID, ok := r.Context().Value(partyWSClientIDContextKey).(int64)
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
						WithField("room_id", roomID).
						WithField("ws_client_id", client.ID())

					for {
						event, recvErr := stream.Recv()
						if recvErr != nil {
							if !errors.Is(recvErr, io.EOF) && !errors.Is(recvErr, context.Canceled) {
								requestLogger.WithField("error", recvErr).Error("party websocket stream closed with error")
							}

							_ = client.Close()

							return
						}

						payload, marshalErr := json.Marshal(event)
						if marshalErr != nil {
							requestLogger.WithField("error", marshalErr).Error("failed to marshal party websocket event")

							_ = client.Close()

							return
						}

						if sendErr := client.Send(payload); sendErr != nil {
							requestLogger.WithField("error", sendErr).Error("failed to send party websocket event")

							_ = client.Close()

							return
						}
					}
				}()

				return nil
			},
			OnClose: func(ctx context.Context, client *wspkg.Client, err error) {
				if cancelValue, ok := streamCancels.LoadAndDelete(client.ID()); ok {
					cancelFn, castOK := cancelValue.(context.CancelFunc)
					if castOK {
						cancelFn()
					}
				}

				if err != nil {
					logger.FromContext(ctx).
						WithField("room_id", roomID).
						WithField("ws_client_id", client.ID()).
						WithField("error", err).
						Error("party websocket closed with error")
				}
			},
		})

		handler(w, r)
	}
}

func wsAuthorization(r *http.Request) string {
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
