package http

import (
	"net/http"
	"strconv"

	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
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

func (h *Handler) GetAllSelections(w http.ResponseWriter, r *http.Request) {
	selections, err := h.usecase.GetAllSelections(r.Context())

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

	selection, err := h.usecase.GetSelectionByTitle(r.Context(), title)
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

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid movie id")
		return
	}

	movie, err := h.usecase.GetMovieByID(r.Context(), id)
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

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid actor id")
		return
	}

	actor, err := h.usecase.GetActorByID(r.Context(), id)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)
		return
	}

	httppkg.Response(w, http.StatusOK, actor)
}

func (h *Handler) GetEpisodePlayback(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	if len(idParam) == 0 {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid episode id")
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid episode id")
		return
	}

	playback, err := h.usecase.GetEpisodePlayback(r.Context(), id, 0)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)
		return
	}

	httppkg.Response(w, http.StatusOK, playback)
}

func (h *Handler) GetEpisodeProgress(w http.ResponseWriter, r *http.Request) {
	auth, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		status, message := errors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)
		return
	}

	idParam := r.PathValue("id")
	if len(idParam) == 0 {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid episode id")
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid episode id")
		return
	}

	progress, err := h.usecase.GetEpisodeProgress(r.Context(), auth.UserId, id)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)
		return
	}

	httppkg.Response(w, http.StatusOK, progress)
}

func (h *Handler) SaveEpisodeProgress(w http.ResponseWriter, r *http.Request) {
	auth, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		status, message := errors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)
		return
	}

	idParam := r.PathValue("id")
	if len(idParam) == 0 {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid episode id")
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid episode id")
		return
	}

	var req domain.WatchProgressRequest
	if err = httppkg.Read(r, &req); err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)
		return
	}

	err = h.usecase.SaveEpisodeProgress(r.Context(), auth.UserId, id, req.PositionSeconds)
	if err != nil {
		status, message := errors.MapError(err)
		httppkg.ErrResponse(w, status, message)
		return
	}

	httppkg.Response(w, http.StatusOK, domain.WatchProgressResponse{
		EpisodeID:       id,
		PositionSeconds: req.PositionSeconds,
	})
}
