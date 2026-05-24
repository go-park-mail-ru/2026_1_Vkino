package routes

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	dto "github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"google.golang.org/grpc"
)

const (
	maxProfileMultipartSize       = 10 << 20
	defaultSearchLimit            = 10
	defaultContinueWatchingLimit  = 5
	defaultCollectionLimit        = 50
	recentWatchHistoryMinProgress = 0.95
	jsonKeySuccess                = "success"
)

type UserClient interface {
	userv1.UserServiceClient
	supportRPC
	movieRPC

	GetSubscriptionCapabilities(
		ctx context.Context,
		in *userv1.GetSubscriptionCapabilitiesRequest,
		opts ...grpc.CallOption,
	) (*userv1.GetSubscriptionCapabilitiesResponse, error)
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
	moviev1.MovieServiceClient

	user userv1.UserServiceClient
	sup  supportRPC
}

func NewUserClient(userConn, movieConn grpc.ClientConnInterface) UserClient {
	return grpcUserClient{
		user:               userv1.NewUserServiceClient(userConn),
		sup:                supportv1.NewSupportServiceClient(userConn),
		MovieServiceClient: moviev1.NewMovieServiceClient(movieConn),
	}
}

func (c grpcUserClient) GetProfile(
	ctx context.Context,
	in *userv1.GetProfileRequest,
	opts ...grpc.CallOption,
) (*userv1.GetProfileResponse, error) {
	return c.user.GetProfile(ctx, in, opts...)
}

func (c grpcUserClient) GetFriend(
	ctx context.Context,
	in *userv1.GetFriendRequest,
	opts ...grpc.CallOption,
) (*userv1.GetFriendResponse, error) {
	return c.user.GetFriend(ctx, in, opts...)
}

func (c grpcUserClient) GetSubscriptionCapabilities(
	ctx context.Context,
	in *userv1.GetSubscriptionCapabilitiesRequest,
	opts ...grpc.CallOption,
) (*userv1.GetSubscriptionCapabilitiesResponse, error) {
	return c.user.GetSubscriptionCapabilities(ctx, in, opts...)
}

func (c grpcUserClient) ActivateSubscription(
	ctx context.Context,
	in *userv1.ActivateSubscriptionRequest,
	opts ...grpc.CallOption,
) (*userv1.ActivateSubscriptionResponse, error) {
	return c.user.ActivateSubscription(ctx, in, opts...)
}

func (c grpcUserClient) SearchUsersByEmail(
	ctx context.Context,
	in *userv1.SearchUsersByEmailRequest,
	opts ...grpc.CallOption,
) (*userv1.SearchUsersByEmailResponse, error) {
	return c.user.SearchUsersByEmail(ctx, in, opts...)
}

func (c grpcUserClient) UpdateProfile(
	ctx context.Context,
	in *userv1.UpdateProfileRequest,
	opts ...grpc.CallOption,
) (*userv1.UpdateProfileResponse, error) {
	return c.user.UpdateProfile(ctx, in, opts...)
}

func (c grpcUserClient) AddFriend(
	ctx context.Context,
	in *userv1.AddFriendRequest,
	opts ...grpc.CallOption,
) (*userv1.AddFriendResponse, error) {
	return c.user.AddFriend(ctx, in, opts...)
}

func (c grpcUserClient) DeleteFriend(
	ctx context.Context,
	in *userv1.DeleteFriendRequest,
	opts ...grpc.CallOption,
) (*userv1.DeleteFriendResponse, error) {
	return c.user.DeleteFriend(ctx, in, opts...)
}

func (c grpcUserClient) AddMovieToFavorites(
	ctx context.Context,
	in *userv1.AddMovieToFavoritesRequest,
	opts ...grpc.CallOption,
) (*userv1.AddMovieToFavoritesResponse, error) {
	return c.user.AddMovieToFavorites(ctx, in, opts...)
}

func (c grpcUserClient) SetMovieRating(
	ctx context.Context,
	in *userv1.SetMovieRatingRequest,
	opts ...grpc.CallOption,
) (*userv1.SetMovieRatingResponse, error) {
	return c.user.SetMovieRating(ctx, in, opts...)
}

func (c grpcUserClient) SetMovieReview(
	ctx context.Context,
	in *userv1.SetMovieReviewRequest,
	opts ...grpc.CallOption,
) (*userv1.SetMovieReviewResponse, error) {
	return c.user.SetMovieReview(ctx, in, opts...)
}

func (c grpcUserClient) DeleteMovieReview(
	ctx context.Context,
	in *userv1.DeleteMovieReviewRequest,
	opts ...grpc.CallOption,
) (*userv1.DeleteMovieReviewResponse, error) {
	return c.user.DeleteMovieReview(ctx, in, opts...)
}

func (c grpcUserClient) SetReviewReaction(
	ctx context.Context,
	in *userv1.SetReviewReactionRequest,
	opts ...grpc.CallOption,
) (*userv1.SetReviewReactionResponse, error) {
	return c.user.SetReviewReaction(ctx, in, opts...)
}

func (c grpcUserClient) DeleteReviewReaction(
	ctx context.Context,
	in *userv1.DeleteReviewReactionRequest,
	opts ...grpc.CallOption,
) (*userv1.DeleteReviewReactionResponse, error) {
	return c.user.DeleteReviewReaction(ctx, in, opts...)
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

func (c grpcUserClient) CreateTicket(
	ctx context.Context,
	in *supportv1.CreateTicketRequest,
	opts ...grpc.CallOption,
) (*supportv1.TicketResponse, error) {
	return c.sup.CreateTicket(ctx, in, opts...)
}

func (c grpcUserClient) GetTickets(
	ctx context.Context,
	in *supportv1.GetTicketsRequest,
	opts ...grpc.CallOption,
) (*supportv1.TicketsResponse, error) {
	return c.sup.GetTickets(ctx, in, opts...)
}

func (c grpcUserClient) UpdateTicket(
	ctx context.Context,
	in *supportv1.UpdateTicketRequest,
	opts ...grpc.CallOption,
) (*supportv1.TicketResponse, error) {
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

type setMovieRatingRequest struct {
	Rating float64 `json:"rating"`
}

type setMovieReviewRequest struct {
	Rating  *float64 `json:"rating"`
	Comment *string  `json:"comment,omitempty"`
	Message *string  `json:"message,omitempty"`
}

type setReviewReactionRequest struct {
	Reaction string `json:"reaction"`
}

func (r setMovieReviewRequest) reviewComment() *string {
	if r.Message != nil {
		return r.Message
	}

	return r.Comment
}

func readUpdateProfilePayload(w http.ResponseWriter, r *http.Request) (updateProfilePayload, bool) {
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		return readMultipartUpdateProfilePayload(w, r)

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

func readMultipartUpdateProfilePayload(w http.ResponseWriter, r *http.Request) (updateProfilePayload, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxProfileMultipartSize)
	if err := r.ParseMultipartForm(maxProfileMultipartSize); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid multipart form body")

		return updateProfilePayload{}, false
	}

	payload := updateProfilePayload{Birthdate: strings.TrimSpace(r.FormValue("birthdate"))}

	header, ok := firstMultipartFile(r.MultipartForm, "avatar")
	if !ok || shouldIgnoreMultipartAvatarHeader(header) {
		return payload, true
	}

	avatarBytes, contentType, ok := readMultipartAvatarFile(w, header)
	if !ok {
		return updateProfilePayload{}, false
	}

	payload.AvatarContentType = contentType
	if isAvatarReferencePayload(avatarBytes, payload.AvatarContentType) {
		logIgnoredAvatarReference(r.Context(), payload.AvatarContentType, avatarBytes)

		return payload, true
	}

	payload.Avatar = avatarBytes

	return payload, true
}

func readMultipartAvatarFile(w http.ResponseWriter, header *multipart.FileHeader) ([]byte, string, bool) {
	file, err := header.Open()
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid avatar file")

		return nil, "", false
	}

	defer func() {
		_ = file.Close()
	}()

	avatarBytes, err := io.ReadAll(file)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "failed to read avatar file")

		return nil, "", false
	}

	return avatarBytes, header.Header.Get("Content-Type"), true
}

func logIgnoredAvatarReference(ctx context.Context, contentType string, avatarBytes []byte) {
	logger.FromContext(ctx).
		WithField("avatar_content_type", contentType).
		WithField("avatar_size", len(avatarBytes)).
		WithField("avatar_preview", string(bytes.TrimSpace(avatarBytes))).
		Info("ignoring avatar reference payload")
}

func isAvatarReferencePayload(body []byte, contentType string) bool {
	trimmedBody := bytes.TrimSpace(body)
	if len(trimmedBody) == 0 {
		return true
	}

	value := strings.ToLower(string(trimmedBody))

	if value == "null" || value == "undefined" {
		return true
	}

	if strings.HasPrefix(value, "blob:") ||
		strings.HasPrefix(value, "http://") ||
		strings.HasPrefix(value, "https://") {
		return true
	}

	trimmedType := strings.ToLower(strings.TrimSpace(contentType))

	return !strings.HasPrefix(trimmedType, "image/")
}

func firstMultipartFile(form *multipart.Form, field string) (*multipart.FileHeader, bool) {
	if form == nil || form.File == nil {
		return nil, false
	}

	files := form.File[field]
	if len(files) == 0 || files[0] == nil {
		return nil, false
	}

	return files[0], true
}

func shouldIgnoreMultipartAvatarHeader(header *multipart.FileHeader) bool {
	if header == nil {
		return true
	}

	filename := strings.ToLower(strings.TrimSpace(header.Filename))
	contentType := strings.ToLower(strings.TrimSpace(header.Header.Get("Content-Type")))

	if filename == "" || filename == "null" || filename == "undefined" || filename == "blob" {
		return true
	}

	if strings.HasPrefix(contentType, "image/") {
		return false
	}

	return false
}

func User(cfg Config, userClient UserClient) []httpserver.Option {
	return []httpserver.Option{
		route("GET /user/me", newUserProfileHandler(cfg, userClient)),
		route("GET /user/subscription/capabilities", newUserSubscriptionCapabilitiesHandler(cfg, userClient)),
		route("GET /user/search", newUserSearchHandler(cfg, userClient)),
		route("PUT /user/profile", newUserUpdateProfileHandler(cfg, userClient)),
		route("POST /user/friends/{id}", newUserSendFriendRequestHandler(cfg, userClient)),
		route("DELETE /user/friends/{id}", newUserDeleteFriendHandler(cfg, userClient)),
		route("PUT /user/favorites/{id}", newUserToggleFavoriteHandler(cfg, userClient)),
		route("PUT /user/ratings/{id}", newUserSetMovieRatingHandler(cfg, userClient)),
		route("PUT /user/reviews/{id}", newUserSetMovieReviewHandler(cfg, userClient)),
		route("DELETE /user/reviews/{id}", newUserDeleteMovieReviewHandler(cfg, userClient)),
		route("PUT /user/review-reactions/{id}", newUserSetReviewReactionHandler(cfg, userClient)),
		route("DELETE /user/review-reactions/{id}", newUserDeleteReviewReactionHandler(cfg, userClient)),
		route("GET /user/favorites", newUserFavoritesHandler(cfg, userClient)),
		route("GET /user/watch/continue", newUserContinueWatchingHandler(cfg, userClient)),
		route("GET /user/watch/history", newUserWatchHistoryHandler(cfg, userClient, 0)),
		route("GET /user/watch/recent", newUserWatchHistoryHandler(cfg, userClient, recentWatchHistoryMinProgress)),
		route("GET /user/friends/requests", newUserFriendRequestsHandler(cfg, userClient)),
		route("POST /user/friends/requests/{id}/respond", newUserRespondFriendRequestHandler(cfg, userClient)),
		route("DELETE /user/friends/requests/{id}", newUserDeleteOutgoingFriendRequestHandler(cfg, userClient)),
		route("GET /user/friends", newUserFriendsListHandler(cfg, userClient)),
		route("POST /support/tickets", newSupportCreateTicketHandler(cfg, userClient)),
		route("POST /support/files", newSupportFileUploadHandler(cfg, userClient)),
		route("GET /support/files", newSupportFileURLHandler(cfg, userClient)),
		route("GET /support/tickets", newSupportTicketsHandler(cfg, userClient)),
		route("PATCH /support/tickets/{id}", newSupportUpdateTicketHandler(cfg, userClient)),
		route("GET /support/tickets/{id}/messages", newSupportTicketMessagesHandler(cfg, userClient)),
		httpserver.WithRoute("GET /support/tickets/{id}/subscribe", newSupportTicketSubscribeHandler(userClient)),
		route("POST /support/tickets/{id}/messages", newSupportCreateTicketMessageHandler(cfg, userClient)),
		route("GET /support/statistics", newSupportStatisticsHandler(cfg, userClient)),
	}
}

func newUserProfileHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.GetProfile(r.Context(), &userv1.GetProfileRequest{})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newUserSubscriptionCapabilitiesHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.GetSubscriptionCapabilities(
			r.Context(),
			&userv1.GetSubscriptionCapabilitiesRequest{},
		)
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, subscriptionStateFromProto(resp))
	}
}

func newUserSearchHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.SearchUsers(r.Context(), &userv1.SearchUsersRequest{
			Query: r.URL.Query().Get("query"),
			Limit: parseInt32Query(r, "limit", defaultSearchLimit),
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newUserUpdateProfileHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newUserSendFriendRequestHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newUserDeleteFriendHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			jsonKeySuccess: true,
		})
	}
}

func newUserToggleFavoriteHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newUserSetMovieRatingHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movieID, ok := parsePathID(w, r, "invalid movie id")
		if !ok {
			return
		}

		var req setMovieRatingRequest
		if !readJSON(w, r, &req) {
			return
		}

		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.SetMovieRating(r.Context(), &userv1.SetMovieRatingRequest{
			MovieId: movieID,
			Rating:  req.Rating,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newUserSetMovieReviewHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movieID, ok := parsePathID(w, r, "invalid movie id")
		if !ok {
			return
		}

		var req setMovieReviewRequest
		if !readJSON(w, r, &req) {
			return
		}

		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.SetMovieReview(r.Context(), &userv1.SetMovieReviewRequest{
			MovieId: movieID,
			Rating:  req.Rating,
			Comment: req.reviewComment(),
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newUserDeleteMovieReviewHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movieID, ok := parsePathID(w, r, "invalid movie id")
		if !ok {
			return
		}

		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		_, err := userClient.DeleteMovieReview(r.Context(), &userv1.DeleteMovieReviewRequest{
			MovieId: movieID,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, map[string]bool{
			jsonKeySuccess: true,
		})
	}
}

func newUserSetReviewReactionHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, ok := parsePathID(w, r, "invalid review id")
		if !ok {
			return
		}

		var req setReviewReactionRequest
		if !readJSON(w, r, &req) {
			return
		}

		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.SetReviewReaction(r.Context(), &userv1.SetReviewReactionRequest{
			ReviewId: reviewID,
			Reaction: req.Reaction,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newUserDeleteReviewReactionHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, ok := parsePathID(w, r, "invalid review id")
		if !ok {
			return
		}

		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		_, err := userClient.DeleteReviewReaction(r.Context(), &userv1.DeleteReviewReactionRequest{
			ReviewId: reviewID,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, map[string]bool{
			jsonKeySuccess: true,
		})
	}
}

func newUserFavoritesHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancelUser := grpcContext(r, cfg.UserRequestTimeout())
		defer cancelUser()

		favResp, err := userClient.GetFavorites(r.Context(), &userv1.GetFavoritesRequest{
			Limit:  parseInt32Query(r, "limit", defaultSearchLimit),
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
	}
}

func newUserContinueWatchingHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.MovieRequestTimeout())
		defer cancel()

		resp, err := userClient.GetContinueWatching(r.Context(), &moviev1.GetContinueWatchingRequest{
			Limit: parseInt32Query(r, "limit", defaultContinueWatchingLimit),
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newUserWatchHistoryHandler(cfg Config, userClient UserClient, minProgress float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.MovieRequestTimeout())
		defer cancel()

		resp, err := userClient.GetWatchHistory(r.Context(), &moviev1.GetWatchHistoryRequest{
			Limit:       parseInt32Query(r, "limit", defaultSearchLimit),
			MinProgress: minProgress,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newUserFriendRequestsHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.GetFriendRequests(r.Context(), &userv1.GetFriendRequestsRequest{
			Direction: r.URL.Query().Get("direction"),
			Limit:     parseInt32Query(r, "limit", defaultCollectionLimit),
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newUserRespondFriendRequestHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newUserDeleteOutgoingFriendRequestHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			jsonKeySuccess: true,
		})
	}
}

func newUserFriendsListHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.GetFriendsList(r.Context(), &userv1.GetFriendsListRequest{
			Limit:  parseInt32Query(r, "limit", defaultCollectionLimit),
			Offset: parseInt32Query(r, "offset", 0),
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newSupportCreateTicketHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newSupportTicketsHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, ok := readSupportTicketsRequest(w, r)
		if !ok {
			return
		}

		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.GetTickets(r.Context(), request)
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func newSupportUpdateTicketHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newSupportTicketMessagesHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newSupportCreateTicketMessageHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func newSupportStatisticsHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.GetTicketStatistics(r.Context(), &supportv1.GetTicketStatisticsRequest{})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func readSupportTicketsRequest(w http.ResponseWriter, r *http.Request) (*supportv1.GetTicketsRequest, bool) {
	query := r.URL.Query()
	supportLine := int64(0)

	if rawSupportLine := strings.TrimSpace(query.Get("support_line")); rawSupportLine != "" {
		parsedSupportLine, err := strconv.ParseInt(rawSupportLine, 10, 64)
		if err != nil {
			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid support line")

			return nil, false
		}

		supportLine = parsedSupportLine
	}

	return &supportv1.GetTicketsRequest{
		Status:      strings.TrimSpace(query.Get("status")),
		Category:    strings.TrimSpace(query.Get("category")),
		UserEmail:   strings.TrimSpace(query.Get("user_email")),
		SupportLine: supportLine,
	}, true
}
