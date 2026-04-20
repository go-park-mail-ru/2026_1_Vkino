package http

import (
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	userusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	pkgerrors "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/errors"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

func (h *Handler) SearchUsersByEmail(w http.ResponseWriter, r *http.Request) {
	auth, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		status, message := pkgerrors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)

		return
	}

	users, err := h.usecase.SearchUsersByEmail(r.Context(), auth.UserId, r.URL.Query().Get("email"))
	if err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusOK, users)
}

func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) {
	auth, friendID, ok := h.friendRequestContext(w, r)
	if !ok {
		return
	}

	friend, err := h.usecase.AddFriend(r.Context(), auth.UserId, friendID)
	if err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusCreated, friend)
}

func (h *Handler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	auth, friendID, ok := h.friendRequestContext(w, r)
	if !ok {
		return
	}

	if err := h.usecase.DeleteFriend(r.Context(), auth.UserId, friendID); err != nil {
		status, message := pkgerrors.MapError(err)
		httppkg.ErrResponse(w, status, message)

		return
	}

	httppkg.Response(w, http.StatusOK, domain.Response{Message: "friend deleted"})
}

func (h *Handler) friendRequestContext(
	w http.ResponseWriter,
	r *http.Request,
) (auth userusecase.AuthContext, friendID int64, ok bool) {
	auth, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		status, message := pkgerrors.MapError(middleware.ErrMidlware)
		httppkg.ErrResponse(w, status, message)

		return userusecase.AuthContext{}, 0, false
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid friend id")

		return userusecase.AuthContext{}, 0, false
	}

	parsedID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || parsedID <= 0 {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid friend id")

		return userusecase.AuthContext{}, 0, false
	}

	return auth, parsedID, true
}
