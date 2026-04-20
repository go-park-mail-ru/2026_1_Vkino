package http

import (
	"net/http"

	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	movieusecase "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/usecase/movie"
)

type MovieHandler struct {
	facade *movieusecase.Facade
}

func NewMovieHandler(facade *movieusecase.Facade) *MovieHandler {
	return &MovieHandler{facade: facade}
}

func (h *MovieHandler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, err := h.facade.GetMovieByID(r.Context(), r)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func (h *MovieHandler) GetActorByID(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, err := h.facade.GetActorByID(r.Context(), r)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func (h *MovieHandler) GetSelectionByTitle(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, err := h.facade.GetSelectionByTitle(r.Context(), r)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func (h *MovieHandler) GetAllSelections(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, err := h.facade.GetAllSelections(r.Context())
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}
