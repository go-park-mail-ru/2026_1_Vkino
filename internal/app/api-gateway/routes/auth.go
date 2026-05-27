package routes

import (
	"net/http"

	dto "github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/domain"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

//go:generate go run -mod=mod github.com/mailru/easyjson/easyjson auth.go

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type authAccessTokenResponse struct {
	//nolint:gosec // This is an HTTP response field name, not stored secret material.
	AccessToken string `json:"access_token"`
}

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type authMessageResponse struct {
	Message string `json:"message"`
}

func Auth(
	cfg Config,
	authClient authv1.AuthServiceClient,
) []httpserver.Option {
	return []httpserver.Option{
		route("POST /user/sign-up", newSignUpHandler(cfg, authClient)),
		route("POST /user/sign-in", newSignInHandler(cfg, authClient)),
		route("POST /user/refresh", newRefreshHandler(cfg, authClient)),
		route("POST /user/logout", newLogoutHandler(cfg, authClient)),
		route("POST /user/change-password", newChangePasswordHandler(cfg, authClient)),
	}
}

func newSignUpHandler(cfg Config, authClient authv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SignUpRequest
		if !readJSON(w, r, &req) {
			return
		}

		cancel := grpcContext(r, cfg.AuthRequestTimeout())
		defer cancel()

		resp, err := authClient.SignUp(r.Context(), &authv1.SignUpRequest{
			Email:    req.Email,
			Password: req.Password(),
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		writeAuthCookie(w, cfg, resp.GetRefreshToken(), false)
		httppkg.Response(w, http.StatusCreated, authAccessTokenResponse{AccessToken: resp.GetAccessToken()})
	}
}

func newSignInHandler(cfg Config, authClient authv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SignInRequest
		if !readJSON(w, r, &req) {
			return
		}

		cancel := grpcContext(r, cfg.AuthRequestTimeout())
		defer cancel()

		resp, err := authClient.SignIn(r.Context(), &authv1.SignInRequest{
			Email:    req.Email,
			Password: req.Password(),
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		writeAuthCookie(w, cfg, resp.GetRefreshToken(), false)
		httppkg.Response(w, http.StatusOK, authAccessTokenResponse{AccessToken: resp.GetAccessToken()})
	}
}

func newRefreshHandler(cfg Config, authClient authv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cfg.RefreshCookieName())
		if err != nil {
			httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		cancel := grpcContext(r, cfg.AuthRequestTimeout())
		defer cancel()

		resp, err := authClient.Refresh(r.Context(), &authv1.RefreshRequest{
			RefreshToken: cookie.Value,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		writeAuthCookie(w, cfg, resp.GetRefreshToken(), false)
		httppkg.Response(w, http.StatusOK, authAccessTokenResponse{AccessToken: resp.GetAccessToken()})
	}
}

func newLogoutHandler(cfg Config, authClient authv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.AuthRequestTimeout())
		defer cancel()

		if _, err := authClient.Logout(r.Context(), &authv1.LogoutRequest{}); err != nil {
			writeGRPCError(w, err)

			return
		}

		writeAuthCookie(w, cfg, "", true)
		httppkg.Response(w, http.StatusOK, authMessageResponse{Message: "successfully log out"})
	}
}

func newChangePasswordHandler(cfg Config, authClient authv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.ChangePasswordRequest
		if !readJSON(w, r, &req) {
			return
		}

		cancel := grpcContext(r, cfg.AuthRequestTimeout())
		defer cancel()

		if _, err := authClient.ChangePassword(r.Context(), &authv1.ChangePasswordRequest{
			OldPassword: req.OldPassword,
			NewPassword: req.NewPassword,
		}); err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, authMessageResponse{Message: "password updated"})
	}
}

func writeAuthCookie(w http.ResponseWriter, cfg Config, refreshToken string, expired bool) {
	var cookie http.Cookie // #nosec G124 -- HttpOnly, Secure, and SameSite are set before SetCookie.

	cookie.Name = cfg.RefreshCookieName()
	cookie.Value = refreshToken
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = cfg.CookieSecure()
	cookie.SameSite = http.SameSiteLaxMode

	if expired {
		cookie.MaxAge = -1
	}

	http.SetCookie(w, &cookie)
}
