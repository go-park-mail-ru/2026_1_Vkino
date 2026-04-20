package http

import (
	"net/http"

	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"
	userusecase "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/usecase/user"
)

type UserHandler struct {
	facade *userusecase.Facade
}

func NewUserHandler(facade *userusecase.Facade) *UserHandler {
	return &UserHandler{facade: facade}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, err := h.facade.GetProfile(r.Context(), r)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateProfileRequest
	if err := httppkg.Read(r, &req); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, statusCode, err := h.facade.UpdateProfile(r.Context(), r, req)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func (h *UserHandler) SearchUsersByEmail(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, err := h.facade.SearchUsersByEmail(r.Context(), r)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func (h *UserHandler) AddFriend(w http.ResponseWriter, r *http.Request) {
	var req dto.AddFriendRequest
	if err := httppkg.Read(r, &req); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, statusCode, err := h.facade.AddFriend(r.Context(), r, req)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func (h *UserHandler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteFriendRequest
	if err := httppkg.Read(r, &req); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, statusCode, err := h.facade.DeleteFriend(r.Context(), r, req)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func (h *UserHandler) AddMovieToFavorites(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, err := h.facade.AddMovieToFavoritesFromPath(r.Context(), r)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}