package routes

import (
	"net/http"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/config"
)

type updateProfileRequest struct {
	Birthdate string `json:"birthdate"`
}

func User(
	cfg *config.Config,
	userClient userv1.UserServiceClient,
	authMiddleware func(http.Handler) http.Handler,
) []httpserver.Option {
	return []httpserver.Option{
		httpserver.WithMiddlewareRoute("GET /user/me", func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := requireAuth(w, r)
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.GetProfile(r.Context(), &userv1.GetProfileRequest{
				UserId: authCtx.UserID,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware),

		httpserver.WithMiddlewareRoute("GET /user/search", func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := requireAuth(w, r)
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.SearchUsersByEmail(r.Context(), &userv1.SearchUsersByEmailRequest{
				UserId:     authCtx.UserID,
				EmailQuery: r.URL.Query().Get("email"),
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware),

		httpserver.WithMiddlewareRoute("PUT /user/profile", func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := requireAuth(w, r)
			if !ok {
				return
			}

			var req updateProfileRequest
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.UpdateProfile(r.Context(), &userv1.UpdateProfileRequest{
				UserId:    authCtx.UserID,
				Birthdate: req.Birthdate,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware),

		httpserver.WithMiddlewareRoute("POST /user/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := requireAuth(w, r)
			if !ok {
				return
			}

			friendID, ok := parsePathID(w, r, "id", "invalid friend id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.AddFriend(r.Context(), &userv1.AddFriendRequest{
				UserId:   authCtx.UserID,
				FriendId: friendID,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware),

		httpserver.WithMiddlewareRoute("DELETE /user/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := requireAuth(w, r)
			if !ok {
				return
			}

			friendID, ok := parsePathID(w, r, "id", "invalid friend id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserGRPC.RequestTimeout)
			defer cancel()

			_, err := userClient.DeleteFriend(r.Context(), &userv1.DeleteFriendRequest{
				UserId:   authCtx.UserID,
				FriendId: friendID,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, map[string]bool{
				"success": true,
			})
		}, authMiddleware),

		httpserver.WithMiddlewareRoute("PUT /user/favorites/{id}", func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := requireAuth(w, r)
			if !ok {
				return
			}

			movieID, ok := parsePathID(w, r, "id", "invalid movie id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.AddMovieToFavorites(r.Context(), &userv1.AddMovieToFavoritesRequest{
				UserId:  authCtx.UserID,
				MovieId: movieID,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware),
	}
}