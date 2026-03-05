package http

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/errors"
	http2 "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(u usecase.Usecase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req domain.SignUpRequest
	if err := http2.Read(r, &req); err != nil {
		http2.ErrResponse(w, http.StatusBadRequest, "invalid json body")

		return
	}

	tokens, err := h.usecase.SignUp(req.Email, req.Password)
	if err != nil {
		status, message := errors.MapError(err)
		http2.ErrResponse(w, status, message)
		
		return
	}

	http.SetCookie(w, h.RefreshCookie(tokens.RefreshToken))

	http2.Response(w, http.StatusCreated, tokens)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req domain.SignInRequest
	if err := http2.Read(r, &req); err != nil {
		http2.ErrResponse(w, http.StatusBadRequest, "invalid json body")

		return
	}

	tokens, err := h.usecase.SignIn(req.Email, req.Password)
	if err != nil {
		status, message := errors.MapError(err)
		http2.ErrResponse(w, status, message)

		return
	}

	http.SetCookie(w, h.RefreshCookie(tokens.RefreshToken))

	http2.Response(w, http.StatusOK, tokens)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cfg := h.usecase.GetConfig()

	cookie, err := r.Cookie(cfg.RefreshCookieName)
	if err != nil {
		http2.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	email, err := h.usecase.ValidateRefreshToken(cookie.Value)
	if err != nil {
		status, message := errors.MapError(err)
		http2.ErrResponse(w, status, message)

		return
	}

	tokenPair, err := h.usecase.Refresh(email)
	if err != nil {
		status, message := errors.MapError(err)
		http2.ErrResponse(w, status, message)

		return
	}

	http.SetCookie(w, h.RefreshCookie(tokenPair.RefreshToken))

	http2.Response(w, http.StatusOK, domain.AccessTokenResponse{
		AccessToken: tokenPair.AccessToken,
	})
}

func (h *Handler) RefreshCookie(refreshToken string) *http.Cookie {
	cfg := h.usecase.GetConfig()

	return &http.Cookie{
		Name:     cfg.RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(cfg.RefreshTokenTTL),
	}

}
