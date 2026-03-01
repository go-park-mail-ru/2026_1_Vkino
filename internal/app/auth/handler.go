package auth

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpjson"
	apperrors "github.com/go-park-mail-ru/2026_1_VKino/internal/app/errors"
)

type Handler struct {
	service *Service
}


func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}


func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /sign-up", h.signUp)
	mux.HandleFunc("POST /sign-in", h.signIn)
	mux.HandleFunc("POST /refresh", h.refresh)
}


// ставим refresh-token в cookie
func setRefreshCookie(h *Handler, w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.service.cfg.RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.service.cfg.CookieSecure, // конфиг! Чтобы можно было пока работать по http на localhost
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(h.service.cfg.RefreshTokenTTL),
	})
}


func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest
	if err := httpjson.ReadJSON(r, &req); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	tokens, err := h.service.SignUp(req.Email, req.Password)
	if err != nil {
		apperrors.WriteServiceError(w, err)
		return
	}
	setRefreshCookie(h, w, tokens.RefreshToken)
	httpjson.WriteJSON(w, http.StatusCreated, tokens)
}


func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest
	if err := httpjson.ReadJSON(r, &req); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	tokens, err := h.service.SignIn(req.Email, req.Password)
	if err != nil {
		apperrors.WriteServiceError(w, err)
		return
	}
	setRefreshCookie(h, w, tokens.RefreshToken)
	httpjson.WriteJSON(w, http.StatusOK, tokens)
}


func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		httpjson.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	email, err := h.service.validateRefreshToken(cookie.Value)
	if err != nil {
		apperrors.WriteServiceError(w, err)
		return
	}
	tokenPair, err := h.service.refresh(email)
	if err != nil {
		apperrors.WriteServiceError(w, err)
		return
	}

	setRefreshCookie(h, w, tokenPair.RefreshToken)
	httpjson.WriteJSON(w, http.StatusOK, accessTokenResponse{
		AccessToken: tokenPair.AccessToken,
	})
}