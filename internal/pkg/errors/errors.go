package errors

import (
	stderrors "errors"
	"net/http"

	userdomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"

	repo "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/inmemory"
)

type HttpErr struct {
	status  int
	message string
}

// мапа внутренняя ошибка -> внешний ответ.
var errToHTTP = map[error]HttpErr{
	userdomain.ErrUserAlreadyExists:  {status: http.StatusConflict, message: "user already exists"},
	userdomain.ErrInvalidCredentials: {status: http.StatusUnauthorized, message: "invalid credentials"},
	userdomain.ErrNoSession:          {status: http.StatusUnauthorized, message: "unauthorized"},
	userdomain.ErrInvalidToken:       {status: http.StatusUnauthorized, message: "unauthorized"},
	userdomain.ErrPasswordMismatch:   {status: http.StatusUnauthorized, message: "invalid credentials"},
	userdomain.ErrInvalidBirthdate:   {status: http.StatusBadRequest, message: "invalid birthdate"},
	userdomain.ErrInvalidAvatar:      {status: http.StatusBadRequest, message: "invalid avatar"},
	userdomain.ErrInternal:           {status: http.StatusInternalServerError, message: "internal server error"},
	storagepkg.ErrInvalidFileType:    {status: http.StatusBadRequest, message: "unsupported file extension"},
	storagepkg.ErrFileTooLarge:       {status: http.StatusBadRequest, message: "file size exceeds the limit"},

	repo.ErrSelectionNotFound: {status: http.StatusNotFound, message: "selection not found"},
	repo.ErrMovieNotFound:     {status: http.StatusNotFound, message: "movie not found"},
	repo.ErrActorNotFound:     {status: http.StatusNotFound, message: "actor not found"},

	http.ErrNoCookie:       {status: http.StatusUnauthorized, message: "unauthorized"},
	middleware.ErrMidlware: {status: http.StatusUnauthorized, message: "unauthorized"},
	httppkg.ErrInvalidJson: {status: http.StatusBadRequest, message: "invalid json body"},
}

func MapError(err error) (int, string) {
	var mappedError HttpErr

	for k, v := range errToHTTP {
		if stderrors.Is(err, k) {
			return v.status, v.message
		}
	}

	mappedError = errToHTTP[userdomain.ErrInternal]

	return mappedError.status, mappedError.message
}
