package routes

import (
	"net/http"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/auth/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/config"
	dto "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/domain"
)


func Auth(
	cfg *config.Config,
	authClient authv1.AuthServiceClient,
	authMiddleware func(http.Handler) http.Handler,
) []httpserver.Option {
	return []httpserver.Option{
		httpserver.WithRoute("POST /user/sign-up", func(w http.ResponseWriter, r *http.Request) {
			var req dto.SignUpRequest
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			resp, err := authClient.SignUp(r.Context(), &authv1.SignUpRequest{
				Email:    req.Email,
				Password: req.Password,
			})
			if err != nil {
				writeGRPCError(w, err)
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
			var req dto.SignInRequest
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			resp, err := authClient.SignIn(r.Context(), &authv1.SignInRequest{
				Email:    req.Email,
				Password: req.Password,
			})
			if err != nil {
				writeGRPCError(w, err)
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

			cancel := grpcContext(r, cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			resp, err := authClient.Refresh(r.Context(), &authv1.RefreshRequest{
				RefreshToken: cookie.Value,
			})
			if err != nil {
				writeGRPCError(w, err)
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
			authCtx, ok := requireAuth(w, r)
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			_, err := authClient.Logout(r.Context(), &authv1.LogoutRequest{
				Email: authCtx.Email,
			})
			if err != nil {
				writeGRPCError(w, err)
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

			httppkg.Response(w, http.StatusOK, map[string]string{
				"message": "successfully log out",
			})
		}, authMiddleware),

		httpserver.WithMiddlewareRoute("POST /user/change-password", func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := requireAuth(w, r)
			if !ok {
				return
			}

			var req dto.ChangePasswordRequest
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.AuthGRPC.RequestTimeout)
			defer cancel()

			_, err := authClient.ChangePassword(r.Context(), &authv1.ChangePasswordRequest{
				UserId:      authCtx.UserID,
				OldPassword: req.OldPassword,
				NewPassword: req.NewPassword,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, map[string]string{
				"message": "password updated",
			})
		}, authMiddleware),
	}
}