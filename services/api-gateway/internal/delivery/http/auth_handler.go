package http

import (
	"net/http"

	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"
	authusecase "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/usecase/auth"
)

type AuthHandler struct {
	facade *authusecase.Facade
}

func NewAuthHandler(facade *authusecase.Facade) *AuthHandler {
	return &AuthHandler{facade: facade}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req dto.SignUpRequest
	if err := httppkg.Read(r, &req); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, cookie, statusCode, err := h.facade.SignUp(r.Context(), req)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	http.SetCookie(w, cookie)
	httppkg.Response(w, statusCode, resp)
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req dto.SignInRequest
	if err := httppkg.Read(r, &req); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, cookie, statusCode, err := h.facade.SignIn(r.Context(), req)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	http.SetCookie(w, cookie)
	httppkg.Response(w, statusCode, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	resp, cookie, statusCode, err := h.facade.Refresh(r.Context(), r)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	http.SetCookie(w, cookie)
	httppkg.Response(w, statusCode, resp)
}

func (h *AuthHandler) LogOut(w http.ResponseWriter, r *http.Request) {
	resp, cookie, statusCode, err := h.facade.LogOut(r.Context(), r)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	http.SetCookie(w, cookie)
	httppkg.Response(w, statusCode, resp)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ChangePasswordRequest
	if err := httppkg.Read(r, &req); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, statusCode, err := h.facade.ChangePassword(r.Context(), r, req)
	if err != nil {
		writeHTTPError(w, statusCode, err)
		return
	}

	httppkg.Response(w, statusCode, resp)
}

func writeHTTPError(w http.ResponseWriter, statusCode int, err error) {
	if err == nil {
		return
	}

	httppkg.ErrResponse(w, statusCode, err.Error())
}
