package routes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	dto "github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/domain"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"google.golang.org/grpc"
)

type UserClient interface {
	userv1.UserServiceClient
	supportRPC
	movieRPC
}

type movieRPC interface {
	GetContinueWatching(ctx context.Context, in *moviev1.GetContinueWatchingRequest, opts ...grpc.CallOption) (
		*moviev1.GetContinueWatchingResponse, error)
	GetWatchHistory(ctx context.Context, in *moviev1.GetWatchHistoryRequest, opts ...grpc.CallOption) (
		*moviev1.GetWatchHistoryResponse, error)
	GetMoviesByIDs(ctx context.Context, in *moviev1.GetMoviesByIDsRequest, opts ...grpc.CallOption) (
		*moviev1.GetMoviesByIDsResponse, error)
}

type supportRPC interface {
	CreateTicket(ctx context.Context, in *supportv1.CreateTicketRequest, opts ...grpc.CallOption) (
		*supportv1.TicketResponse, error)
	GetTickets(ctx context.Context, in *supportv1.GetTicketsRequest, opts ...grpc.CallOption) (
		*supportv1.TicketsResponse, error)
	UpdateTicket(ctx context.Context, in *supportv1.UpdateTicketRequest, opts ...grpc.CallOption) (
		*supportv1.TicketResponse, error)
	UploadSupportFile(ctx context.Context, in *supportv1.UploadSupportFileRequest, opts ...grpc.CallOption) (
		*supportv1.UploadSupportFileResponse, error)
	GetSupportFileURL(ctx context.Context, in *supportv1.GetSupportFileURLRequest, opts ...grpc.CallOption) (
		*supportv1.GetSupportFileURLResponse, error)
	GetTicketMessages(ctx context.Context, in *supportv1.GetTicketMessagesRequest, opts ...grpc.CallOption) (
		*supportv1.TicketMessagesResponse, error)
	CreateTicketMessage(ctx context.Context, in *supportv1.CreateTicketMessageRequest, opts ...grpc.CallOption) (
		*supportv1.TicketMessageResponse, error)
	GetTicketStatistics(ctx context.Context, in *supportv1.GetTicketStatisticsRequest, opts ...grpc.CallOption) (
		*supportv1.TicketStatisticsResponse, error)
	SubscribeTicket(ctx context.Context, in *supportv1.SubscribeTicketRequest, opts ...grpc.CallOption) (
		grpc.ServerStreamingClient[supportv1.TicketEvent], error)
}

type grpcUserClient struct {
	user userv1.UserServiceClient
	sup  supportRPC
	moviev1.MovieServiceClient
}

func NewUserClient(userConn, movieConn grpc.ClientConnInterface) UserClient {
	return grpcUserClient{
		user:               userv1.NewUserServiceClient(userConn),
		sup:                supportv1.NewSupportServiceClient(userConn),
		MovieServiceClient: moviev1.NewMovieServiceClient(movieConn),
	}
}

func (c grpcUserClient) GetProfile(ctx context.Context, in *userv1.GetProfileRequest, opts ...grpc.CallOption) (
	*userv1.GetProfileResponse, error) {
	return c.user.GetProfile(ctx, in, opts...)
}

func (c grpcUserClient) SearchUsersByEmail(
	ctx context.Context,
	in *userv1.SearchUsersByEmailRequest,
	opts ...grpc.CallOption,
) (*userv1.SearchUsersByEmailResponse, error) {
	return c.user.SearchUsersByEmail(ctx, in, opts...)
}

func (c grpcUserClient) UpdateProfile(ctx context.Context, in *userv1.UpdateProfileRequest, opts ...grpc.CallOption) (
	*userv1.UpdateProfileResponse, error) {
	return c.user.UpdateProfile(ctx, in, opts...)
}

func (c grpcUserClient) AddFriend(ctx context.Context, in *userv1.AddFriendRequest, opts ...grpc.CallOption) (
	*userv1.AddFriendResponse, error) {
	return c.user.AddFriend(ctx, in, opts...)
}

func (c grpcUserClient) DeleteFriend(ctx context.Context, in *userv1.DeleteFriendRequest, opts ...grpc.CallOption) (
	*userv1.DeleteFriendResponse, error) {
	return c.user.DeleteFriend(ctx, in, opts...)
}

func (c grpcUserClient) AddMovieToFavorites(
	ctx context.Context,
	in *userv1.AddMovieToFavoritesRequest,
	opts ...grpc.CallOption,
) (*userv1.AddMovieToFavoritesResponse, error) {
	return c.user.AddMovieToFavorites(ctx, in, opts...)
}

func (c grpcUserClient) ToggleFavorite(
	ctx context.Context,
	in *userv1.ToggleFavoriteRequest,
	opts ...grpc.CallOption,
) (*userv1.ToggleFavoriteResponse, error) {
	return c.user.ToggleFavorite(ctx, in, opts...)
}

func (c grpcUserClient) GetFavorites(
	ctx context.Context,
	in *userv1.GetFavoritesRequest,
	opts ...grpc.CallOption,
) (*userv1.GetFavoritesResponse, error) {
	return c.user.GetFavorites(ctx, in, opts...)
}

func (c grpcUserClient) SearchUsers(
	ctx context.Context,
	in *userv1.SearchUsersRequest,
	opts ...grpc.CallOption,
) (*userv1.SearchUsersResponse, error) {
	return c.user.SearchUsers(ctx, in, opts...)
}

func (c grpcUserClient) SendFriendRequest(
	ctx context.Context,
	in *userv1.SendFriendRequestRequest,
	opts ...grpc.CallOption,
) (*userv1.SendFriendRequestResponse, error) {
	return c.user.SendFriendRequest(ctx, in, opts...)
}

func (c grpcUserClient) RespondToFriendRequest(
	ctx context.Context,
	in *userv1.RespondToFriendRequestRequest,
	opts ...grpc.CallOption,
) (*userv1.RespondToFriendRequestResponse, error) {
	return c.user.RespondToFriendRequest(ctx, in, opts...)
}

func (c grpcUserClient) DeleteOutgoingFriendRequest(
	ctx context.Context,
	in *userv1.DeleteOutgoingFriendRequestRequest,
	opts ...grpc.CallOption,
) (*userv1.DeleteOutgoingFriendRequestResponse, error) {
	return c.user.DeleteOutgoingFriendRequest(ctx, in, opts...)
}

func (c grpcUserClient) GetFriendRequests(
	ctx context.Context,
	in *userv1.GetFriendRequestsRequest,
	opts ...grpc.CallOption,
) (*userv1.GetFriendRequestsResponse, error) {
	return c.user.GetFriendRequests(ctx, in, opts...)
}

func (c grpcUserClient) GetFriendsList(
	ctx context.Context,
	in *userv1.GetFriendsListRequest,
	opts ...grpc.CallOption,
) (*userv1.GetFriendsListResponse, error) {
	return c.user.GetFriendsList(ctx, in, opts...)
}

func (c grpcUserClient) CreateTicket(ctx context.Context, in *supportv1.CreateTicketRequest, opts ...grpc.CallOption) (
	*supportv1.TicketResponse, error) {
	return c.sup.CreateTicket(ctx, in, opts...)
}

func (c grpcUserClient) GetTickets(ctx context.Context, in *supportv1.GetTicketsRequest, opts ...grpc.CallOption) (
	*supportv1.TicketsResponse, error) {
	return c.sup.GetTickets(ctx, in, opts...)
}

func (c grpcUserClient) UpdateTicket(ctx context.Context, in *supportv1.UpdateTicketRequest, opts ...grpc.CallOption) (
	*supportv1.TicketResponse, error) {
	return c.sup.UpdateTicket(ctx, in, opts...)
}

func (c grpcUserClient) UploadSupportFile(
	ctx context.Context,
	in *supportv1.UploadSupportFileRequest,
	opts ...grpc.CallOption,
) (*supportv1.UploadSupportFileResponse, error) {
	return c.sup.UploadSupportFile(ctx, in, opts...)
}

func (c grpcUserClient) GetSupportFileURL(
	ctx context.Context,
	in *supportv1.GetSupportFileURLRequest,
	opts ...grpc.CallOption,
) (*supportv1.GetSupportFileURLResponse, error) {
	return c.sup.GetSupportFileURL(ctx, in, opts...)
}

func (c grpcUserClient) GetTicketMessages(
	ctx context.Context,
	in *supportv1.GetTicketMessagesRequest,
	opts ...grpc.CallOption,
) (*supportv1.TicketMessagesResponse, error) {
	return c.sup.GetTicketMessages(ctx, in, opts...)
}

func (c grpcUserClient) CreateTicketMessage(
	ctx context.Context,
	in *supportv1.CreateTicketMessageRequest,
	opts ...grpc.CallOption,
) (*supportv1.TicketMessageResponse, error) {
	return c.sup.CreateTicketMessage(ctx, in, opts...)
}

func (c grpcUserClient) GetTicketStatistics(
	ctx context.Context,
	in *supportv1.GetTicketStatisticsRequest,
	opts ...grpc.CallOption,
) (*supportv1.TicketStatisticsResponse, error) {
	return c.sup.GetTicketStatistics(ctx, in, opts...)
}

func (c grpcUserClient) SubscribeTicket(
	ctx context.Context,
	in *supportv1.SubscribeTicketRequest,
	opts ...grpc.CallOption,
) (grpc.ServerStreamingClient[supportv1.TicketEvent], error) {
	return c.sup.SubscribeTicket(ctx, in, opts...)
}

type updateProfileJSONRequest struct {
	Birthdate string `json:"birthdate"`
}

type updateProfilePayload struct {
	Birthdate         string
	Avatar            []byte
	AvatarContentType string
}

func readUpdateProfilePayload(w http.ResponseWriter, r *http.Request) (updateProfilePayload, bool) {
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		// лимит тела запроса, чтобы не тащить бесконечный файл в память
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid multipart form body")

			return updateProfilePayload{}, false
		}

		payload := updateProfilePayload{
			Birthdate: strings.TrimSpace(r.FormValue("birthdate")),
		}

		file, header, err := r.FormFile("avatar")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				return payload, true
			}

			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid avatar file")

			return updateProfilePayload{}, false
		}

		defer func() {
			_ = file.Close()
		}()

		avatarBytes, err := io.ReadAll(file)
		if err != nil {
			httppkg.ErrResponse(w, http.StatusBadRequest, "failed to read avatar file")

			return updateProfilePayload{}, false
		}

		payload.Avatar = avatarBytes
		if header != nil {
			payload.AvatarContentType = header.Header.Get("Content-Type")
		}

		return payload, true

	default:
		var req updateProfileJSONRequest
		if !readJSON(w, r, &req) {
			return updateProfilePayload{}, false
		}

		return updateProfilePayload{
			Birthdate: strings.TrimSpace(req.Birthdate),
		}, true
	}
}

func User(cfg Config, userClient UserClient) []httpserver.Option {
	return []httpserver.Option{
		route("GET /user/me", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.GetProfile(r.Context(), &userv1.GetProfileRequest{})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /user/search", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.SearchUsers(r.Context(), &userv1.SearchUsersRequest{
				Query: r.URL.Query().Get("query"),
				Limit: parseInt32Query(r, "limit", 10),
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("PUT /user/profile", func(w http.ResponseWriter, r *http.Request) {
			req, ok := readUpdateProfilePayload(w, r)
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.UpdateProfile(r.Context(), &userv1.UpdateProfileRequest{
				Birthdate:         req.Birthdate,
				Avatar:            req.Avatar,
				AvatarContentType: req.AvatarContentType,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("POST /user/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
			toUserID, ok := parsePathID(w, r, "invalid friend id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.SendFriendRequest(r.Context(), &userv1.SendFriendRequestRequest{
				ToUserId: toUserID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("DELETE /user/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
			friendID, ok := parsePathID(w, r, "invalid friend id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			_, err := userClient.DeleteFriend(r.Context(), &userv1.DeleteFriendRequest{
				FriendId: friendID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, map[string]bool{
				"success": true,
			})
		}),

		route("PUT /user/favorites/{id}", func(w http.ResponseWriter, r *http.Request) {
			movieID, ok := parsePathID(w, r, "invalid movie id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.ToggleFavorite(r.Context(), &userv1.ToggleFavoriteRequest{
				MovieId: movieID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /user/favorites", func(w http.ResponseWriter, r *http.Request) {
			cancelUser := grpcContext(r, cfg.UserRequestTimeout())
			defer cancelUser()

			favResp, err := userClient.GetFavorites(r.Context(), &userv1.GetFavoritesRequest{
				Limit:  parseInt32Query(r, "limit", 10),
				Offset: parseInt32Query(r, "offset", 0),
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			movieIDs := favResp.GetMovieIds()

			out := dto.FavoritesHTTPResponse{
				MovieIDs:   movieIDs,
				TotalCount: favResp.GetTotalCount(),
				Movies:     []*moviev1.MovieCard{},
			}

			if len(movieIDs) == 0 {
				httppkg.Response(w, http.StatusOK, out)

				return
			}

			cancelMovie := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancelMovie()

			moviesResp, err := userClient.GetMoviesByIDs(r.Context(), &moviev1.GetMoviesByIDsRequest{
				MovieIds: movieIDs,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			out.Movies = orderMovieCardsByIDOrder(movieIDs, moviesResp.GetMovies())
			httppkg.Response(w, http.StatusOK, out)
		}),

		route("GET /user/watch/continue", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := userClient.GetContinueWatching(r.Context(), &moviev1.GetContinueWatchingRequest{
				Limit: parseInt32Query(r, "limit", 5),
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /user/watch/history", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := userClient.GetWatchHistory(r.Context(), &moviev1.GetWatchHistoryRequest{
				Limit:       parseInt32Query(r, "limit", 10),
				MinProgress: 0,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /user/watch/recent", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := userClient.GetWatchHistory(r.Context(), &moviev1.GetWatchHistoryRequest{
				Limit:       parseInt32Query(r, "limit", 10),
				MinProgress: 0.95,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /user/friends/requests", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.GetFriendRequests(r.Context(), &userv1.GetFriendRequestsRequest{
				Direction: r.URL.Query().Get("direction"),
				Limit:     parseInt32Query(r, "limit", 50),
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("POST /user/friends/requests/{id}/respond", func(w http.ResponseWriter, r *http.Request) {
			requestID, ok := parsePathID(w, r, "invalid request id")
			if !ok {
				return
			}

			var req struct {
				Action string `json:"action"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.RespondToFriendRequest(r.Context(), &userv1.RespondToFriendRequestRequest{
				RequestId: requestID,
				Action:    req.Action,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("DELETE /user/friends/requests/{id}", func(w http.ResponseWriter, r *http.Request) {
			requestID, ok := parsePathID(w, r, "invalid request id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			_, err := userClient.DeleteOutgoingFriendRequest(r.Context(), &userv1.DeleteOutgoingFriendRequestRequest{
				RequestId: requestID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, map[string]bool{
				"success": true,
			})
		}),

		route("GET /user/friends", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.GetFriendsList(r.Context(), &userv1.GetFriendsListRequest{
				Limit:  parseInt32Query(r, "limit", 50),
				Offset: parseInt32Query(r, "offset", 0),
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("POST /support/tickets", func(w http.ResponseWriter, r *http.Request) {
			var req dto.SupportCreateTicketRequest
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.CreateTicket(r.Context(), &supportv1.CreateTicketRequest{
				Category:          req.Category,
				Title:             req.Title,
				Description:       req.Description,
				UserEmail:         strings.TrimSpace(req.UserEmail),
				AttachmentFileKey: req.AttachmentFileKey,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusCreated, resp)
		}),

		route("POST /support/files", newSupportFileUploadHandler(cfg, userClient)),

		route("GET /support/files", newSupportFileURLHandler(cfg, userClient)),

		route("GET /support/tickets", func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()

			supportLine := int64(0)

			if rawSupportLine := strings.TrimSpace(query.Get("support_line")); rawSupportLine != "" {
				parsedSupportLine, err := strconv.ParseInt(rawSupportLine, 10, 64)
				if err != nil {
					httppkg.ErrResponse(w, http.StatusBadRequest, "invalid support line")

					return
				}

				supportLine = parsedSupportLine
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			request := &supportv1.GetTicketsRequest{
				Status:      strings.TrimSpace(query.Get("status")),
				Category:    strings.TrimSpace(query.Get("category")),
				UserEmail:   strings.TrimSpace(query.Get("user_email")),
				SupportLine: supportLine,
			}

			resp, err := userClient.GetTickets(r.Context(), request)
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("PATCH /support/tickets/{id}", func(w http.ResponseWriter, r *http.Request) {
			ticketID, ok := parsePathID(w, r, "invalid ticket id")
			if !ok {
				return
			}

			var req dto.SupportUpdateTicketRequest
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.UpdateTicket(r.Context(), &supportv1.UpdateTicketRequest{
				TicketId:          ticketID,
				Category:          req.Category,
				Status:            req.Status,
				SupportLine:       req.SupportLine,
				Title:             req.Title,
				UserEmail:         strings.TrimSpace(req.UserEmail),
				Description:       req.Description,
				AttachmentFileKey: req.AttachmentFileKey,
				Rating:            req.Rating,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /support/tickets/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
			ticketID, ok := parsePathID(w, r, "invalid ticket id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.GetTicketMessages(r.Context(), &supportv1.GetTicketMessagesRequest{
				TicketId: ticketID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /support/tickets/{id}/subscribe", newSupportTicketSubscribeHandler(userClient)),

		route("POST /support/tickets/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
			ticketID, ok := parsePathID(w, r, "invalid ticket id")
			if !ok {
				return
			}

			var req dto.SupportCreateTicketMessageRequest
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.CreateTicketMessage(r.Context(), &supportv1.CreateTicketMessageRequest{
				TicketId:       ticketID,
				Content:        req.Content,
				ContentFileKey: req.ContentFileKey,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusCreated, resp)
		}),

		route("GET /support/statistics", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.GetTicketStatistics(r.Context(), &supportv1.GetTicketStatisticsRequest{})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),
	}
}
