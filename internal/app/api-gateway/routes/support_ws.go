package routes

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
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

		authorization, ok := authorizeSubscription(w, r)
		if !ok {
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

		serveSubscription(w, wsSubscribeConfig[supportv1.TicketEvent]{
			hubKey:        ticketID,
			resourceName:  "support",
			resourceField: "ticket_id",
			clientIDKey:   supportWSClientIDContextKey,
			request:       r,
			stream:        stream,
			cancelStream:  cancelStream,
			streamCancels: streamCancels,
			nextClientID:  &nextClientID,
			hubs:          hubs,
			streamErrLog:  "support websocket stream closed with error",
			marshalErrLog: "failed to marshal support websocket event",
			sendErrLog:    "failed to send support websocket event",
			closeErrLog:   "support websocket closed with error",
		})
	}
}

func supportWSAuthorization(r *http.Request) string {
	return subscribeAuthorization(r)
}
