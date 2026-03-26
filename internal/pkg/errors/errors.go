package errors

import (
	stderrors "errors"
	"net/http"

	authdomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"

	repo "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/inmemory"
)

type HttpErr struct {
	status  int
	message string
}

// мапа внутренняя ошибка -> внешний ответ.
var errToHTTP = map[error]HttpErr{
	authdomain.ErrUserAlreadyExists:  {status: http.StatusConflict, message: "user already exists"},
	authdomain.ErrInvalidCredentials: {status: http.StatusUnauthorized, message: "invalid credentials"},
	authdomain.ErrNoSession:          {status: http.StatusUnauthorized, message: "unauthorized"},
	authdomain.ErrInvalidToken:       {status: http.StatusUnauthorized, message: "unauthorized"},
	authdomain.ErrInternal:           {status: http.StatusInternalServerError, message: "internal server error"},

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

	mappedError = errToHTTP[authdomain.ErrInternal]

	return mappedError.status, mappedError.message
}
