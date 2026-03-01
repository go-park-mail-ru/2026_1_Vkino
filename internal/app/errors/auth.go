package apperrors

import (
	stderrors "errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpjson"
)

var (
	ErrUserAlreadyExists  = stderrors.New("user already exists")
	ErrInvalidCredentials = stderrors.New("invalid credentials")
	ErrNoSession          = stderrors.New("no session")
	ErrInvalidToken       = stderrors.New("invalid token")
)

type httpErr struct {
	status  int
	message string
}

// Мапа "внутренняя ошибка -> внешний ответ"
var errToHTTP = map[error]httpErr{
	ErrUserAlreadyExists:  {status: http.StatusConflict, message: "user already exists"},
	ErrInvalidCredentials: {status: http.StatusUnauthorized, message: "invalid credentials"},
	ErrNoSession:          {status: http.StatusUnauthorized, message: "unauthorized"},
	ErrInvalidToken:       {status: http.StatusUnauthorized, message: "unauthorized"},
}

// перекладывать статус и тело в writeError
func WriteServiceError(w http.ResponseWriter, err error) {
	var key error

	switch {
	case stderrors.Is(err, ErrUserAlreadyExists):
		key = ErrUserAlreadyExists
	case stderrors.Is(err, ErrInvalidCredentials):
		key = ErrInvalidCredentials
	case stderrors.Is(err, ErrNoSession):
		key = ErrNoSession
	case stderrors.Is(err, ErrInvalidToken):
		key = ErrInvalidToken
	default:
		httpjson.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	mapped := errToHTTP[key]
	httpjson.WriteError(w, mapped.status, mapped.message)
}