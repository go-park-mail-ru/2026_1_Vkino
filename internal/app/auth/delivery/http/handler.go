package http

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/errors"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(u usecase.Usecase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req domain.SignUpRequest
	if err := httppkg.Read(r, &req); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")

		return
	}

	tokens, err := h.usecase.SignUp(req.Email, req.Password)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	http.SetCookie(w, h.RefreshCookie(tokens.RefreshToken))

	httppkg.Response(w, http.StatusCreated, tokens)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req domain.SignInRequest
	if err := httppkg.Read(r, &req); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")

		return
	}

	tokens, err := h.usecase.SignIn(req.Email, req.Password)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	http.SetCookie(w, h.RefreshCookie(tokens.RefreshToken))

	httppkg.Response(w, http.StatusOK, tokens)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cfg := h.usecase.GetConfig()

	cookie, err := r.Cookie(cfg.RefreshCookieName)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	email, err := h.usecase.ValidateRefreshToken(cookie.Value)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	tokenPair, err := h.usecase.Refresh(email)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	http.SetCookie(w, h.RefreshCookie(tokenPair.RefreshToken))

	httppkg.Response(w, http.StatusOK, domain.AccessTokenResponse{
		AccessToken: tokenPair.AccessToken,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	email, ok := middleware.UserEmailFromContext(r.Context())

	if !ok {
		httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	httppkg.Response(w, http.StatusOK, domain.Response{
		Email: email,
	})
}

func (h *Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	cfg := h.usecase.GetConfig()

	email, ok := middleware.UserEmailFromContext(r.Context())

	if !ok {
		httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

		// тут не нужна проверка, потому что через мидлвар прошло
		// для отладки оставила, чтобы если что поймать ошибку
		return
	}

	err := h.usecase.LogOut(email)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)
		return
	}

	deletedCookie := &http.Cookie{
		Name:     cfg.RefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, deletedCookie)

	httppkg.Response(w, http.StatusOK, domain.Response{
		Message: "successfully log out",
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
