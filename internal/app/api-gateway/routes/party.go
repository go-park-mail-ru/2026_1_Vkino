//nolint:gocognit // HTTP route registration remains intentionally flat for readability.
package routes

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"google.golang.org/grpc"
)

type PartyClient interface {
	GetOverview(ctx context.Context, in *partyv1.GetOverviewRequest, opts ...grpc.CallOption) (*partyv1.GetOverviewResponse, error)
	GetRoom(ctx context.Context, in *partyv1.GetRoomRequest, opts ...grpc.CallOption) (*partyv1.GetRoomResponse, error)
	CreateRoom(ctx context.Context, in *partyv1.CreateRoomRequest, opts ...grpc.CallOption) (*partyv1.CreateRoomResponse, error)
	JoinRoom(ctx context.Context, in *partyv1.JoinRoomRequest, opts ...grpc.CallOption) (*partyv1.JoinRoomResponse, error)
	DeleteRoom(ctx context.Context, in *partyv1.DeleteRoomRequest, opts ...grpc.CallOption) (*partyv1.DeleteRoomResponse, error)
	SubscribeRoom(ctx context.Context, in *partyv1.SubscribeRoomRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[partyv1.RoomEvent], error)
}

//nolint:gocyclo,cyclop // Route registration intentionally stays flat for readability.
func Party(cfg Config, partyClient PartyClient) []httpserver.Option {
	return []httpserver.Option{
		route("GET /watch-party/overview", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.GetOverview(r.Context(), &partyv1.GetOverviewRequest{})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /watch-party/rooms/{id}", func(w http.ResponseWriter, r *http.Request) {
			roomID, ok := parseRoomPathID(w, r, "invalid room id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.GetRoom(r.Context(), &partyv1.GetRoomRequest{
				RoomId: roomID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("POST /watch-party/rooms", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Name       string `json:"name"`
				Visibility string `json:"visibility"`
				MovieID    int64  `json:"movie_id"`
				EpisodeID  int64  `json:"episode_id"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.CreateRoom(r.Context(), &partyv1.CreateRoomRequest{
				Name:       req.Name,
				Visibility: req.Visibility,
				MovieId:    req.MovieID,
				EpisodeId:  req.EpisodeID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusCreated, resp)
		}),

		route("POST /watch-party/join", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				InviteLink string `json:"invite_link"`
				RoomID     int64  `json:"room_id"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.JoinRoom(r.Context(), &partyv1.JoinRoomRequest{
				InviteLink: req.InviteLink,
				RoomId:     req.RoomID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("DELETE /watch-party/rooms/{id}", func(w http.ResponseWriter, r *http.Request) {
			roomID, ok := parseRoomPathID(w, r, "invalid room id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.DeleteRoom(r.Context(), &partyv1.DeleteRoomRequest{
				RoomId: roomID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /watch-party/rooms/{id}/subscribe", newPartyRoomSubscribeHandler(partyClient)),
	}
}

func parseRoomPathID(w http.ResponseWriter, r *http.Request, message string) (int64, bool) {
	value := strings.TrimSpace(r.PathValue("id"))
	if value == "" {
		httppkg.ErrResponse(w, http.StatusBadRequest, message)

		return 0, false
	}

	value = strings.TrimPrefix(strings.ToLower(value), "id")

	roomID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, message)

		return 0, false
	}

	return roomID, true
}
