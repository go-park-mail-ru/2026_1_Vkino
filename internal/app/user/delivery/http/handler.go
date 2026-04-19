package http

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	pkgerrors "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/errors"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
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
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	tokens, err := h.usecase.SignUp(r.Context(), req.Email, req.Password)
	if err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	http.SetCookie(w, h.RefreshCookie(tokens.RefreshToken))

	httppkg.Response(w, http.StatusCreated, tokens)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req domain.SignInRequest
	if err := httppkg.Read(r, &req); err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	tokens, err := h.usecase.SignIn(r.Context(), req.Email, req.Password)
	if err != nil {
		status, message := pkgerrors.MapError(err)
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
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	email, err := h.usecase.ValidateRefreshToken(r.Context(), cookie.Value)
	if err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	tokenPair, err := h.usecase.Refresh(r.Context(), email)
	if err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	http.SetCookie(w, h.RefreshCookie(tokenPair.RefreshToken))

	httppkg.Response(w, http.StatusOK, domain.AccessTokenResponse{
		AccessToken: tokenPair.AccessToken,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	auth, err := middleware.AuthFromContext(r.Context())

	if err != nil {
		status, message := pkgerrors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)

		return
	}

	profile, err := h.usecase.GetProfile(r.Context(), auth.UserId)
	if err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	auth, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		status, message := pkgerrors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)

		return
	}

	requestLogger := logger.FromContext(r.Context()).
		WithField("handler", "user.UpdateProfile")

	const maxAvatarSize = 5 * 1024 * 1024 // 5 МБ
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		status, message := pkgerrors.MapError(domain.ErrInvalidAvatar)
		httppkg.ErrResponse(w, status, message)

		return
	}

	birthdate := strings.TrimSpace(r.FormValue("birthdate"))

	var (
		file        multipart.File
		header      *multipart.FileHeader
		contentType string
		size        int64
	)

	file, header, err = r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		size = header.Size
		if size <= 0 {
			status, message := pkgerrors.MapError(domain.ErrInvalidAvatar)
			httppkg.ErrResponse(w, status, message)

			return
		}

		if size > maxAvatarSize {
			status, message := pkgerrors.MapError(storagepkg.ErrFileTooLarge)
			httppkg.ErrResponse(w, status, message)

			return
		}

		contentType = header.Header.Get("Content-Type")

		requestLogger.
			WithField("size", size).
			WithField("content_type", contentType).
			Info("avatar received")
	} else if !errors.Is(err, http.ErrMissingFile) {
		status, message := pkgerrors.MapError(domain.ErrInvalidAvatar)
		httppkg.ErrResponse(w, status, message)

		return
	}

	profile, err := h.usecase.UpdateProfile(r.Context(), auth.UserId, birthdate, file, size, contentType)
	if err != nil {
		requestLogger.
			WithField("birthdate", birthdate).
			WithField("has_avatar", file != nil).
			WithField("error", err).
			Error("update profile failed")
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusOK, profile)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	auth, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		status, message := pkgerrors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)

		return
	}

	var req domain.ChangePasswordRequest
	if err := httppkg.Read(r, &req); err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	err = h.usecase.ChangePassword(r.Context(), auth.UserId, req.OldPassword, req.NewPassword)
	if err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusOK, domain.Response{Message: "password updated"})
}

func (h *Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	cfg := h.usecase.GetConfig()

	auth, err := middleware.AuthFromContext(r.Context())

	if err != nil {
		status, message := pkgerrors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)

		return
	}

	err = h.usecase.LogOut(r.Context(), auth.Email)
	if err != nil {
		status, message := pkgerrors.MapError(err)
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
