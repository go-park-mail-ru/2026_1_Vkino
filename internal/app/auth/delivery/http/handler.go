package http

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/errors"
)

type Handler struct {
	usecase *usecase.AuthUsecase
}

func NewHandler(u *usecase.AuthUsecase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /sign-up", h.SignUp)
	mux.HandleFunc("POST /sign-in", h.SignIn)
	mux.HandleFunc("POST /refresh", h.Refresh)
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req domain.SignUpRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")

		return
	}

	tokens, err := h.usecase.SignUp(req.Email, req.Password)
	if err != nil {
		errors.WriteServiceError(w, err)

		return
	}

	h.setRefreshCookie(w, tokens.RefreshToken)

	err = WriteJSON(w, http.StatusCreated, tokens)
	if err != nil {
		errors.WriteServiceError(w, err)
	}
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req domain.SignInRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")

		return
	}

	tokens, err := h.usecase.SignIn(req.Email, req.Password)
	if err != nil {
		errors.WriteServiceError(w, err)

		return
	}

	h.setRefreshCookie(w, tokens.RefreshToken)

	err = WriteJSON(w, http.StatusOK, tokens)
	if err != nil {
		errors.WriteServiceError(w, err)
	}
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cfg := h.usecase.GetConfig()

	cookie, err := r.Cookie(cfg.RefreshCookieName)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	email, err := h.usecase.ValidateRefreshToken(cookie.Value)
	if err != nil {
		errors.WriteServiceError(w, err)

		return
	}

	tokenPair, err := h.usecase.Refresh(email)
	if err != nil {
		errors.WriteServiceError(w, err)

		return
	}

	h.setRefreshCookie(w, tokenPair.RefreshToken)

	err = WriteJSON(w, http.StatusOK, domain.AccessTokenResponse{
		AccessToken: tokenPair.AccessToken,
	})
	if err != nil {
		errors.WriteServiceError(w, err)
	}
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, refreshToken string) {
	cfg := h.usecase.GetConfig()

	http.SetCookie(w, &http.Cookie{
		Name:     cfg.RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(cfg.RefreshTokenTTL),
	})
}
