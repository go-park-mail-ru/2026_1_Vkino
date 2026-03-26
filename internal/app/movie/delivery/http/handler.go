package http

import (
	"net/http"

	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/errors"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/google/uuid"
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

func (h *Handler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	if len(idParam) == 0 {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid movie id")
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid movie id")
		return
	}

	movie, err := h.usecase.GetMovieByID(id)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)
		return
	}

	httppkg.Response(w, http.StatusOK, movie)
}

func (h *Handler) GetActorByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	if len(idParam) == 0 {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid actor id")
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid actor id")
		return
	}

	actor, err := h.usecase.GetActorByID(id)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)
		return
	}

	httppkg.Response(w, http.StatusOK, actor)
}