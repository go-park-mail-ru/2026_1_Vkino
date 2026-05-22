package routes

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
	wspkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/ws"
)

type partyWSContextKey string

const partyWSClientIDContextKey partyWSContextKey = "party_ws_client_id"

func newPartyRoomSubscribeHandler(partyClient PartyClient) http.HandlerFunc {
	var nextClientID atomic.Int64

	hubs := wspkg.NewGroup()
	streamCancels := &sync.Map{}

	return func(w http.ResponseWriter, r *http.Request) {
		roomID, ok := parseRoomPathID(w, r)
		if !ok {
			return
		}

		authorization, ok := authorizeSubscription(w, r)
		if !ok {
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

		serveSubscription(w, wsSubscribeConfig[partyv1.RoomEvent]{
			hubKey:        roomID,
			resourceName:  "party",
			resourceField: "room_id",
			clientIDKey:   partyWSClientIDContextKey,
			request:       r,
			stream:        stream,
			cancelStream:  cancelStream,
			streamCancels: streamCancels,
			nextClientID:  &nextClientID,
			hubs:          hubs,
			streamErrLog:  "party websocket stream closed with error",
			marshalErrLog: "failed to marshal party websocket event",
			sendErrLog:    "failed to send party websocket event",
			closeErrLog:   "party websocket closed with error",
		})
	}
}
