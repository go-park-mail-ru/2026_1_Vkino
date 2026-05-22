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
	PartyOverviewClient
	PartyRoomClient
	PartyRealtimeClient
}

type PartyOverviewClient interface {
	GetOverview(ctx context.Context, in *partyv1.GetOverviewRequest,
		opts ...grpc.CallOption) (*partyv1.GetOverviewResponse, error)
}

type PartyRoomClient interface {
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
}

type PartyRealtimeClient interface {
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

func Party(
	cfg Config,
	partyClient PartyClient,
	movieClient moviev1.MovieServiceClient,
	friendClient PartyFriendClient,
) []httpserver.Option {
	return []httpserver.Option{
		route("GET /watch-party/overview", newPartyOverviewHandler(cfg, partyClient, movieClient)),
		route("GET /watch-party/rooms/{id}", newPartyRoomHandler(cfg, partyClient)),
		route("POST /watch-party/rooms", newCreatePartyRoomHandler(cfg, partyClient)),
		route("POST /watch-party/rooms/{id}/invite", newPartyInviteHandler(cfg, partyClient)),
		route("POST /watch-party/friends/{friendId}/invite", newInviteFriendToPartyHandler(cfg, partyClient, friendClient)),
		route("POST /watch-party/join", newJoinPartyHandler(cfg, partyClient)),
		route("GET /watch-party/join/{inviteCode}", newJoinPartyByCodeHandler(cfg, partyClient)),
		route("DELETE /watch-party/rooms/{id}", newDeletePartyRoomHandler(cfg, partyClient)),
		route("POST /watch-party/rooms/{id}/actions", newPartyActionHandler(cfg, partyClient)),
		route("POST /watch-party/rooms/{id}/messages", newPartyMessageHandler(cfg, partyClient)),
		route("POST /watch-party/rooms/{id}/polls", newPartyPollHandler(cfg, partyClient)),
		route("POST /watch-party/rooms/{id}/polls/{pollId}/votes", newPartyVoteHandler(cfg, partyClient)),

		httpserver.WithRoute("GET /watch-party/rooms/{id}/subscribe", newPartyRoomSubscribeHandler(partyClient)),
	}
}

func newPartyOverviewHandler(
	cfg Config,
	partyClient PartyClient,
	movieClient moviev1.MovieServiceClient,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newPartyRoomHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID, ok := parseRoomPathID(w, r, "invalid room id")
		if !ok {
			return
		}

		cancel := grpcContext(r, cfg.PartyRequestTimeout())
		defer cancel()

		resp, err := partyClient.GetRoom(r.Context(), &partyv1.GetRoomRequest{RoomId: roomID})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newCreatePartyRoomHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			Name: req.Name, Visibility: req.Visibility, MovieId: req.MovieID, EpisodeId: req.EpisodeID,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusCreated, resp)
	}
}

func newPartyInviteHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID, ok := parseRoomPathID(w, r, "invalid room id")
		if !ok {
			return
		}

		cancel := grpcContext(r, cfg.PartyRequestTimeout())
		defer cancel()

		resp, err := partyClient.GetRoomInvite(r.Context(), &partyv1.GetRoomInviteRequest{RoomId: roomID})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newInviteFriendToPartyHandler(
	cfg Config,
	partyClient PartyClient,
	friendClient PartyFriendClient,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		friend, err := friendClient.GetFriend(r.Context(), &userv1.GetFriendRequest{FriendId: friendID})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		cancelParty := grpcContext(r, cfg.PartyRequestTimeout())
		defer cancelParty()

		resp, err := partyClient.InviteFriendToRoom(r.Context(), &partyv1.InviteFriendToRoomRequest{
			RoomId: req.RoomID, InvitedUserId: friendID,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, domain.PartyFriendInviteHTTP{
			RoomID: resp.GetRoomId(), Status: resp.GetStatus(), Friend: friend,
		})
	}
}

func newJoinPartyHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			InviteLink string `json:"invite_link"`
		}
		if !readJSON(w, r, &req) {
			return
		}

		joinParty(w, r, cfg, partyClient, req.InviteLink)
	}
}

func newJoinPartyByCodeHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inviteCode := strings.TrimSpace(r.PathValue("inviteCode"))
		if inviteCode == "" {
			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid invite link")

			return
		}

		joinParty(w, r, cfg, partyClient, inviteCode)
	}
}

func joinParty(w http.ResponseWriter, r *http.Request, cfg Config, partyClient PartyClient, inviteLink string) {
	cancel := grpcContext(r, cfg.PartyRequestTimeout())
	defer cancel()

	resp, err := partyClient.JoinRoom(r.Context(), &partyv1.JoinRoomRequest{InviteLink: inviteLink})
	if err != nil {
		writeGRPCError(w, err)

		return
	}

	httppkg.Response(w, http.StatusOK, resp)
}

func newDeletePartyRoomHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID, ok := parseRoomPathID(w, r, "invalid room id")
		if !ok {
			return
		}

		cancel := grpcContext(r, cfg.PartyRequestTimeout())
		defer cancel()

		resp, err := partyClient.DeleteRoom(r.Context(), &partyv1.DeleteRoomRequest{RoomId: roomID})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newPartyActionHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			RoomId: roomID, Action: req.Action, MovieId: req.MovieID, EpisodeId: req.EpisodeID,
			PlaybackUrl: req.PlaybackURL, DurationSeconds: req.DurationSeconds,
			PositionSeconds: req.PositionSeconds, Status: req.Status,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newPartyMessageHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			RoomId: roomID, Content: req.Content,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusCreated, resp)
	}
}

func newPartyPollHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			RoomId: roomID, Question: req.Question, Options: req.Options,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusCreated, resp)
	}
}

func newPartyVoteHandler(cfg Config, partyClient PartyClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			RoomId: roomID, PollId: pollID, OptionId: req.OptionID,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}
