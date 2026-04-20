package gateway

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/movie/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/config"
	authmw "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http/middleware"
)

type AuthMiddleware interface {
	Middleware(next http.Handler) http.Handler
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type updateProfileRequest struct {
	Birthdate string `json:"birthdate"`
}

type addFriendRequest struct {
	FriendID int64 `json:"friend_id"`
}

type deleteFriendRequest struct {
	FriendID int64 `json:"friend_id"`
}

func RegisterRoutes(
	cfg *config.Config,
	authClient authv1.AuthServiceClient,
	userClient userv1.UserServiceClient,
	movieClient moviev1.MovieServiceClient,
	authMiddleware AuthMiddleware,
) []httpserver.Option {
	return []httpserver.Option{
		httpserver.WithRoute("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),

		httpserver.WithRoute("POST /user/sign-up", func(w http.ResponseWriter, r *http.Request) {
			var req signUpRequest
			if err := httppkg.Read(r, &req); err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			resp, err := authClient.SignUp(ctx, &authv1.SignUpRequest{
				Email:    req.Email,
				Password: req.Password,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     cfg.UserAuth.RefreshCookieName,
				Value:    resp.GetRefreshToken(),
				Path:     "/",
				HttpOnly: true,
				Secure:   cfg.UserAuth.CookieSecure,
				SameSite: http.SameSiteLaxMode,
			})

			httppkg.Response(w, http.StatusCreated, map[string]string{
				"access_token": resp.GetAccessToken(),
			})
		}),

		httpserver.WithRoute("POST /user/sign-in", func(w http.ResponseWriter, r *http.Request) {
			var req signInRequest
			if err := httppkg.Read(r, &req); err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			resp, err := authClient.SignIn(ctx, &authv1.SignInRequest{
				Email:    req.Email,
				Password: req.Password,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     cfg.UserAuth.RefreshCookieName,
				Value:    resp.GetRefreshToken(),
				Path:     "/",
				HttpOnly: true,
				Secure:   cfg.UserAuth.CookieSecure,
				SameSite: http.SameSiteLaxMode,
			})

			httppkg.Response(w, http.StatusOK, map[string]string{
				"access_token": resp.GetAccessToken(),
			})
		}),

		httpserver.WithRoute("POST /user/refresh", func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.UserAuth.RefreshCookieName)
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			resp, err := authClient.Refresh(ctx, &authv1.RefreshRequest{
				RefreshToken: cookie.Value,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     cfg.UserAuth.RefreshCookieName,
				Value:    resp.GetRefreshToken(),
				Path:     "/",
				HttpOnly: true,
				Secure:   cfg.UserAuth.CookieSecure,
				SameSite: http.SameSiteLaxMode,
			})

			httppkg.Response(w, http.StatusOK, map[string]string{
				"access_token": resp.GetAccessToken(),
			})
		}),

		httpserver.WithMiddlewareRoute("POST /user/logout", func(w http.ResponseWriter, r *http.Request) {
			authCtx, err := authmw.AuthFromContext(r.Context())
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			_, err = authClient.Logout(ctx, &authv1.LogoutRequest{
				Email: authCtx.Email,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     cfg.UserAuth.RefreshCookieName,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   cfg.UserAuth.CookieSecure,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			})

			httppkg.Response(w, http.StatusOK, map[string]string{"message": "successfully log out"})
		}, authMiddleware.Middleware),

		httpserver.WithMiddlewareRoute("POST /user/change-password", func(w http.ResponseWriter, r *http.Request) {
			authCtx, err := authmw.AuthFromContext(r.Context())
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			var req changePasswordRequest
			if err := httppkg.Read(r, &req); err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			_, err = authClient.ChangePassword(ctx, &authv1.ChangePasswordRequest{
				UserId:      authCtx.UserID,
				OldPassword: req.OldPassword,
				NewPassword: req.NewPassword,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, map[string]string{"message": "password updated"})
		}, authMiddleware.Middleware),

		httpserver.WithMiddlewareRoute("GET /user/me", func(w http.ResponseWriter, r *http.Request) {
			authCtx, err := authmw.AuthFromContext(r.Context())
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.GetProfile(ctx, &userv1.GetProfileRequest{
				UserId: authCtx.UserID,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware.Middleware),

		httpserver.WithMiddlewareRoute("PUT /user/profile", func(w http.ResponseWriter, r *http.Request) {
			authCtx, err := authmw.AuthFromContext(r.Context())
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			var req updateProfileRequest
			if err := httppkg.Read(r, &req); err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.UpdateProfile(ctx, &userv1.UpdateProfileRequest{
				UserId:    authCtx.UserID,
				Birthdate: req.Birthdate,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware.Middleware),

		httpserver.WithMiddlewareRoute("GET /user/search", func(w http.ResponseWriter, r *http.Request) {
			authCtx, err := authmw.AuthFromContext(r.Context())
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.SearchUsersByEmail(ctx, &userv1.SearchUsersByEmailRequest{
				UserId:     authCtx.UserID,
				EmailQuery: r.URL.Query().Get("email"),
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware.Middleware),

		httpserver.WithMiddlewareRoute("POST /user/friend", func(w http.ResponseWriter, r *http.Request) {
			authCtx, err := authmw.AuthFromContext(r.Context())
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			var req addFriendRequest
			if err := httppkg.Read(r, &req); err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.AddFriend(ctx, &userv1.AddFriendRequest{
				UserId:   authCtx.UserID,
				FriendId: req.FriendID,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware.Middleware),

		httpserver.WithMiddlewareRoute("DELETE /user/friend", func(w http.ResponseWriter, r *http.Request) {
			authCtx, err := authmw.AuthFromContext(r.Context())
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			var req deleteFriendRequest
			if err := httppkg.Read(r, &req); err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.UserGRPC.RequestTimeout)
			defer cancel()

			_, err = userClient.DeleteFriend(ctx, &userv1.DeleteFriendRequest{
				UserId:   authCtx.UserID,
				FriendId: req.FriendID,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, map[string]bool{"success": true})
		}, authMiddleware.Middleware),

		httpserver.WithMiddlewareRoute("POST /movie/{id}/favorite", func(w http.ResponseWriter, r *http.Request) {
			authCtx, err := authmw.AuthFromContext(r.Context())
			if err != nil {
				httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			movieID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
			if err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid movie id")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.UserGRPC.RequestTimeout)
			defer cancel()

			resp, err := userClient.AddMovieToFavorites(ctx, &userv1.AddMovieToFavoritesRequest{
				UserId:  authCtx.UserID,
				MovieId: movieID,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}, authMiddleware.Middleware),

		httpserver.WithRoute("GET /movie/selection/all", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.MovieGRPC.RequestTimeout)
			defer cancel()

			resp, err := movieClient.GetAllSelections(ctx, &moviev1.GetAllSelectionsRequest{})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /movie/selection/{selection}", func(w http.ResponseWriter, r *http.Request) {
			title := strings.TrimSpace(r.PathValue("selection"))
			if title == "" {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid selection title")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.MovieGRPC.RequestTimeout)
			defer cancel()

			resp, err := movieClient.GetSelectionByTitle(ctx, &moviev1.GetSelectionByTitleRequest{
				Title: title,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /movie/{id}", func(w http.ResponseWriter, r *http.Request) {
			movieID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
			if err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid movie id")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.MovieGRPC.RequestTimeout)
			defer cancel()

			resp, err := movieClient.GetMovieByID(ctx, &moviev1.GetMovieByIDRequest{
				MovieId: movieID,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /movie/actor/{id}", func(w http.ResponseWriter, r *http.Request) {
			actorID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
			if err != nil {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid actor id")
				return
			}

			ctx, cancel := grpcx.WithTimeout(r.Context(), cfg.MovieGRPC.RequestTimeout)
			defer cancel()

			resp, err := movieClient.GetActorByID(ctx, &moviev1.GetActorByIDRequest{
				ActorId: actorID,
			})
			if err != nil {
				grpcx.WriteError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),
	}
}
