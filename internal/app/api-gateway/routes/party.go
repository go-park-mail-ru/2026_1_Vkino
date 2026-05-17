//nolint:gocognit // HTTP route registration remains intentionally flat for readability.
package routes

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/domain"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"google.golang.org/grpc"
)

type PartyClient interface {
	GetOverview(ctx context.Context, in *partyv1.GetOverviewRequest,
		opts ...grpc.CallOption) (*partyv1.GetOverviewResponse, error)
	GetRoom(ctx context.Context, in *partyv1.GetRoomRequest,
		opts ...grpc.CallOption) (*partyv1.GetRoomResponse, error)
	GetRoomInvite(ctx context.Context, in *partyv1.GetRoomInviteRequest,
		opts ...grpc.CallOption) (*partyv1.GetRoomInviteResponse, error)
	InviteFriendToRoom(ctx context.Context, in *partyv1.InviteFriendToRoomRequest,
		opts ...grpc.CallOption) (*partyv1.InviteFriendToRoomResponse, error)
	CreateRoom(ctx context.Context, in *partyv1.CreateRoomRequest,
		opts ...grpc.CallOption) (*partyv1.CreateRoomResponse, error)
	JoinRoom(ctx context.Context, in *partyv1.JoinRoomRequest,
		opts ...grpc.CallOption) (*partyv1.JoinRoomResponse, error)
	DeleteRoom(ctx context.Context, in *partyv1.DeleteRoomRequest,
		opts ...grpc.CallOption) (*partyv1.DeleteRoomResponse, error)
	ApplyRoomAction(ctx context.Context, in *partyv1.ApplyRoomActionRequest,
		opts ...grpc.CallOption) (*partyv1.ApplyRoomActionResponse, error)
	SendRoomMessage(ctx context.Context, in *partyv1.SendRoomMessageRequest,
		opts ...grpc.CallOption) (*partyv1.SendRoomMessageResponse, error)
	CreateRoomPoll(ctx context.Context, in *partyv1.CreateRoomPollRequest,
		opts ...grpc.CallOption) (*partyv1.CreateRoomPollResponse, error)
	VoteRoomPoll(ctx context.Context, in *partyv1.VoteRoomPollRequest,
		opts ...grpc.CallOption) (*partyv1.VoteRoomPollResponse, error)
	SubscribeRoom(ctx context.Context, in *partyv1.SubscribeRoomRequest,
		opts ...grpc.CallOption) (grpc.ServerStreamingClient[partyv1.RoomEvent], error)
}

type PartyFriendClient interface {
	GetFriend(ctx context.Context, in *userv1.GetFriendRequest, opts ...grpc.CallOption) (*userv1.GetFriendResponse, error)
}

//nolint:gocyclo,cyclop // Route registration intentionally stays flat for readability.
func Party(
	cfg Config,
	partyClient PartyClient,
	movieClient moviev1.MovieServiceClient,
	friendClient PartyFriendClient,
) []httpserver.Option {
	return []httpserver.Option{
		route("GET /watch-party/overview", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.GetOverview(r.Context(), &partyv1.GetOverviewRequest{})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			movieImages, err := loadMovieImageURLs(r.Context(), movieClient, collectOverviewMovieIDs(resp))
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, mapOverviewResponse(resp, movieImages))
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

		route("POST /watch-party/rooms/{id}/invite", func(w http.ResponseWriter, r *http.Request) {
			roomID, ok := parseRoomPathID(w, r, "invalid room id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.GetRoomInvite(r.Context(), &partyv1.GetRoomInviteRequest{
				RoomId: roomID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("POST /watch-party/friends/{friendId}/invite", func(w http.ResponseWriter, r *http.Request) {
			friendID, ok := parseNamedPathID(w, r, "friendId", "invalid friend id")
			if !ok {
				return
			}

			var req struct {
				RoomID int64 `json:"room_id"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancelUser := grpcContext(r, cfg.UserRequestTimeout())
			defer cancelUser()

			friend, err := friendClient.GetFriend(r.Context(), &userv1.GetFriendRequest{
				FriendId: friendID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			cancelParty := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancelParty()

			resp, err := partyClient.InviteFriendToRoom(r.Context(), &partyv1.InviteFriendToRoomRequest{
				RoomId:        req.RoomID,
				InvitedUserId: friendID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, domain.PartyFriendInviteHTTP{
				RoomID: resp.GetRoomId(),
				Status: resp.GetStatus(),
				Friend: friend,
			})
		}),

		route("POST /watch-party/join", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				InviteLink string `json:"invite_link"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.JoinRoom(r.Context(), &partyv1.JoinRoomRequest{
				InviteLink: req.InviteLink,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /watch-party/join/{inviteCode}", func(w http.ResponseWriter, r *http.Request) {
			inviteCode := strings.TrimSpace(r.PathValue("inviteCode"))
			if inviteCode == "" {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid invite link")

				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.JoinRoom(r.Context(), &partyv1.JoinRoomRequest{
				InviteLink: inviteCode,
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

		route("POST /watch-party/rooms/{id}/actions", func(w http.ResponseWriter, r *http.Request) {
			roomID, ok := parseRoomPathID(w, r, "invalid room id")
			if !ok {
				return
			}

			var req struct {
				Action          string `json:"action"`
				MovieID         int64  `json:"movie_id"`
				EpisodeID       int64  `json:"episode_id"`
				PlaybackURL     string `json:"playback_url"`
				DurationSeconds int64  `json:"duration_seconds"`
				PositionSeconds int64  `json:"position_seconds"`
				Status          string `json:"status"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.ApplyRoomAction(r.Context(), &partyv1.ApplyRoomActionRequest{
				RoomId:          roomID,
				Action:          req.Action,
				MovieId:         req.MovieID,
				EpisodeId:       req.EpisodeID,
				PlaybackUrl:     req.PlaybackURL,
				DurationSeconds: req.DurationSeconds,
				PositionSeconds: req.PositionSeconds,
				Status:          req.Status,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("POST /watch-party/rooms/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
			roomID, ok := parseRoomPathID(w, r, "invalid room id")
			if !ok {
				return
			}

			var req struct {
				Content string `json:"content"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.SendRoomMessage(r.Context(), &partyv1.SendRoomMessageRequest{
				RoomId:  roomID,
				Content: req.Content,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusCreated, resp)
		}),

		route("POST /watch-party/rooms/{id}/polls", func(w http.ResponseWriter, r *http.Request) {
			roomID, ok := parseRoomPathID(w, r, "invalid room id")
			if !ok {
				return
			}

			var req struct {
				Question string   `json:"question"`
				Options  []string `json:"options"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.CreateRoomPoll(r.Context(), &partyv1.CreateRoomPollRequest{
				RoomId:   roomID,
				Question: req.Question,
				Options:  req.Options,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusCreated, resp)
		}),

		route("POST /watch-party/rooms/{id}/polls/{pollId}/votes", func(w http.ResponseWriter, r *http.Request) {
			roomID, ok := parseRoomPathID(w, r, "invalid room id")
			if !ok {
				return
			}

			pollID, ok := parseNamedPathID(w, r, "pollId", "invalid poll id")
			if !ok {
				return
			}

			var req struct {
				OptionID int64 `json:"option_id"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.PartyRequestTimeout())
			defer cancel()

			resp, err := partyClient.VoteRoomPoll(r.Context(), &partyv1.VoteRoomPollRequest{
				RoomId:   roomID,
				PollId:   pollID,
				OptionId: req.OptionID,
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
