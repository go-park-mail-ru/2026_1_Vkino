package errors

import (
	stderrors "errors"
	"net/http"

	authdomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
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
