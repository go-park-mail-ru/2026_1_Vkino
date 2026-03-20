package http

import (
	"net/http"

	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/errors"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(u usecase.Usecase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) GetAllSelections(w http.ResponseWriter, r *http.Request) {
	selections, err := h.usecase.GetAllSelections()

	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusOK, selections)
}

func (h *Handler) GetSelectionByTitle(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/movie/selection/")

	if len(title) == 0 {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid selection")

		return
	}

	selection, err := h.usecase.GetSelectionByTitle(title)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusOK, selection)
}
