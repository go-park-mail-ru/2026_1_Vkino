package routes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	dto "github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/domain"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"google.golang.org/grpc"
)

type UserClient interface {
	userv1.UserServiceClient
	supportRPC
}

type supportRPC interface {
	CreateTicket(ctx context.Context, in *supportv1.CreateTicketRequest, opts ...grpc.CallOption) (
		*supportv1.TicketResponse, error)
	GetTickets(ctx context.Context, in *supportv1.GetTicketsRequest, opts ...grpc.CallOption) (
		*supportv1.TicketsResponse, error)
	UpdateTicket(ctx context.Context, in *supportv1.UpdateTicketRequest, opts ...grpc.CallOption) (
		*supportv1.TicketResponse, error)
	GetTicketMessages(ctx context.Context, in *supportv1.GetTicketMessagesRequest, opts ...grpc.CallOption) (
		*supportv1.TicketMessagesResponse, error)
	CreateTicketMessage(ctx context.Context, in *supportv1.CreateTicketMessageRequest, opts ...grpc.CallOption) (
		*supportv1.TicketMessageResponse, error)
	GetTicketStatistics(ctx context.Context, in *supportv1.GetTicketStatisticsRequest, opts ...grpc.CallOption) (
		*supportv1.TicketStatisticsResponse, error)
}

type grpcUserClient struct {
	user userv1.UserServiceClient
	sup  supportRPC
}

func NewUserClient(conn grpc.ClientConnInterface) UserClient {
	return grpcUserClient{
		user: userv1.NewUserServiceClient(conn),
		sup:  supportv1.NewSupportServiceClient(conn),
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

func User(
	cfg Config,
	userClient UserClient,
) []httpserver.Option {
	return []httpserver.Option{
		httpserver.WithRoute("GET /user/me", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.GetProfile(r.Context(), &userv1.GetProfileRequest{})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /user/search", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.SearchUsersByEmail(r.Context(), &userv1.SearchUsersByEmailRequest{
				EmailQuery: r.URL.Query().Get("email"),
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("PUT /user/profile", func(w http.ResponseWriter, r *http.Request) {
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

		httpserver.WithRoute("POST /user/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
			friendID, ok := parsePathID(w, r, "invalid friend id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.AddFriend(r.Context(), &userv1.AddFriendRequest{
				FriendId: friendID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("DELETE /user/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
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

		httpserver.WithRoute("PUT /user/favorites/{id}", func(w http.ResponseWriter, r *http.Request) {
			movieID, ok := parsePathID(w, r, "invalid movie id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.AddMovieToFavorites(r.Context(), &userv1.AddMovieToFavoritesRequest{
				MovieId: movieID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("POST /support/tickets", func(w http.ResponseWriter, r *http.Request) {
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

		httpserver.WithRoute("GET /support/tickets", func(w http.ResponseWriter, r *http.Request) {
			var req dto.SupportGetTicketsRequest
			if !readJSON(w, r, &req) {
				return
			}

			role := strings.TrimSpace(req.Role)
			if role == "" {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid role")

				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			request := &supportv1.GetTicketsRequest{
				Status:      strings.TrimSpace(req.Status),
				Category:    strings.TrimSpace(req.Category),
				UserEmail:   strings.TrimSpace(req.UserEmail),
				SupportLine: req.SupportLine,
			}

			switch role {
			case "user", "support_l1", "support_l2", "admin":
			default:
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid role")

				return
			}

			resp, err := userClient.GetTickets(r.Context(), request)
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("PATCH /support/tickets/{id}", func(w http.ResponseWriter, r *http.Request) {
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
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /support/tickets/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
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

		httpserver.WithRoute("POST /support/tickets/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
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

		httpserver.WithRoute("GET /support/statistics", func(w http.ResponseWriter, r *http.Request) {
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
