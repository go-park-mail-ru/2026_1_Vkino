package http

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/usecase"
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

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	auth, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		status, message := errors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)

		return
	}

	profile, err := h.usecase.GetProfile(r.Context(), auth.UserId)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusOK, profile)
}
