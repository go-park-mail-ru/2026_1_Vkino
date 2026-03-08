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

func (h *Handler) GetSelectionByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httppkg.ErrResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	title := strings.TrimPrefix(r.URL.Path, "/lists/movies/")

	if title == "" {
		httppkg.ErrResponse(w, http.StatusBadRequest, "selection is required")
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
