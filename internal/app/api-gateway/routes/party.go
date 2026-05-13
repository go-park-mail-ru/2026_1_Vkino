//nolint:gocognit // HTTP route registration remains intentionally flat for readability.
package routes

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
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
	ApplyRoomAction(ctx context.Context, in *partyv1.ApplyRoomActionRequest, opts ...grpc.CallOption) (*partyv1.ApplyRoomActionResponse, error)
	SendRoomMessage(ctx context.Context, in *partyv1.SendRoomMessageRequest, opts ...grpc.CallOption) (*partyv1.SendRoomMessageResponse, error)
	CreateRoomPoll(ctx context.Context, in *partyv1.CreateRoomPollRequest, opts ...grpc.CallOption) (*partyv1.CreateRoomPollResponse, error)
	VoteRoomPoll(ctx context.Context, in *partyv1.VoteRoomPollRequest, opts ...grpc.CallOption) (*partyv1.VoteRoomPollResponse, error)
	SubscribeRoom(ctx context.Context, in *partyv1.SubscribeRoomRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[partyv1.RoomEvent], error)
}

type partyOverviewResponse struct {
	ActiveRooms   []partyRoomCardHTTP `json:"active_rooms"`
	MyRooms       []partyRoomCardHTTP `json:"my_rooms"`
	FeaturedRooms []partyRoomCardHTTP `json:"featured_rooms"`
}

type partyRoomCardHTTP struct {
	ID                int64                   `json:"id"`
	Name              string                  `json:"name"`
	Visibility        string                  `json:"visibility"`
	InviteLink        string                  `json:"invite_link"`
	HostUserID        int64                   `json:"host_user_id"`
	HostName          string                  `json:"host_name"`
	ParticipantsCount int32                   `json:"participants_count"`
	Playback          *partyPlaybackStateHTTP `json:"playback,omitempty"`
	UpdatedAt         string                  `json:"updated_at"`
}

type partyPlaybackStateHTTP struct {
	MovieID         int64   `json:"movie_id"`
	MovieTitle      *string `json:"movie_title"`
	EpisodeID       int64   `json:"episode_id"`
	ImgURL          *string `json:"img_url"`
	PlaybackURL     string  `json:"playback_url,omitempty"`
	DurationSeconds int64   `json:"duration_seconds"`
	PositionSeconds int64   `json:"position_seconds"`
	Status          string  `json:"status"`
	UpdatedAt       string  `json:"updated_at"`
}

//nolint:gocyclo,cyclop // Route registration intentionally stays flat for readability.
func Party(cfg Config, partyClient PartyClient, movieClient moviev1.MovieServiceClient) []httpserver.Option {
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

func collectOverviewMovieIDs(resp *partyv1.GetOverviewResponse) []int64 {
	seen := make(map[int64]struct{})
	result := make([]int64, 0)

	appendFromRooms := func(items []*partyv1.RoomCard) {
		for _, item := range items {
			if item == nil || item.GetPlayback() == nil {
				continue
			}

			movieID := item.GetPlayback().GetMovieId()
			if movieID == 0 {
				continue
			}

			if _, ok := seen[movieID]; ok {
				continue
			}

			seen[movieID] = struct{}{}
			result = append(result, movieID)
		}
	}

	appendFromRooms(resp.GetActiveRooms())
	appendFromRooms(resp.GetMyRooms())
	appendFromRooms(resp.GetFeaturedRooms())

	return result
}

func loadMovieImageURLs(
	ctx context.Context,
	movieClient moviev1.MovieServiceClient,
	movieIDs []int64,
) (map[int64]movieOverviewMeta, error) {
	result := make(map[int64]movieOverviewMeta, len(movieIDs))
	if len(movieIDs) == 0 {
		return result, nil
	}

	resp, err := movieClient.GetMoviesByIDs(ctx, &moviev1.GetMoviesByIDsRequest{MovieIds: movieIDs})
	if err != nil {
		return nil, err
	}

	for _, movie := range resp.GetMovies() {
		if movie == nil {
			continue
		}

		meta := movieOverviewMeta{}

		if value := strings.TrimSpace(movie.GetTitle()); value != "" {
			valueCopy := value
			meta.title = &valueCopy
		}

		if value := strings.TrimSpace(movie.GetImgUrl()); value != "" {
			valueCopy := value
			meta.imageURL = &valueCopy
		}

		result[movie.GetId()] = meta
	}

	return result, nil
}

type movieOverviewMeta struct {
	title    *string
	imageURL *string
}

func mapOverviewResponse(resp *partyv1.GetOverviewResponse, movieImages map[int64]movieOverviewMeta) partyOverviewResponse {
	return partyOverviewResponse{
		ActiveRooms:   mapRoomCardsHTTP(resp.GetActiveRooms(), movieImages),
		MyRooms:       mapRoomCardsHTTP(resp.GetMyRooms(), movieImages),
		FeaturedRooms: mapRoomCardsHTTP(resp.GetFeaturedRooms(), movieImages),
	}
}

func mapRoomCardsHTTP(items []*partyv1.RoomCard, movieImages map[int64]movieOverviewMeta) []partyRoomCardHTTP {
	result := make([]partyRoomCardHTTP, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		result = append(result, partyRoomCardHTTP{
			ID:                item.GetId(),
			Name:              item.GetName(),
			Visibility:        item.GetVisibility(),
			InviteLink:        item.GetInviteLink(),
			HostUserID:        item.GetHostUserId(),
			HostName:          item.GetHostName(),
			ParticipantsCount: item.GetParticipantsCount(),
			Playback:          mapPlaybackHTTP(item.GetPlayback(), movieImages),
			UpdatedAt:         item.GetUpdatedAt(),
		})
	}

	return result
}

func mapPlaybackHTTP(item *partyv1.PlaybackState, movieImages map[int64]movieOverviewMeta) *partyPlaybackStateHTTP {
	if item == nil {
		return nil
	}

	meta := movieImages[item.GetMovieId()]

	return &partyPlaybackStateHTTP{
		MovieID:         item.GetMovieId(),
		MovieTitle:      meta.title,
		EpisodeID:       item.GetEpisodeId(),
		ImgURL:          meta.imageURL,
		PlaybackURL:     item.GetPlaybackUrl(),
		DurationSeconds: item.GetDurationSeconds(),
		PositionSeconds: item.GetPositionSeconds(),
		Status:          item.GetStatus(),
		UpdatedAt:       item.GetUpdatedAt(),
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

func parseNamedPathID(w http.ResponseWriter, r *http.Request, name, message string) (int64, bool) {
	value := strings.TrimSpace(r.PathValue(name))
	if value == "" {
		httppkg.ErrResponse(w, http.StatusBadRequest, message)

		return 0, false
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, message)

		return 0, false
	}

	return id, true
}
