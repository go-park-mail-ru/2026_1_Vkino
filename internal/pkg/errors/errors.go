package errors

import (
	"encoding/json"
	stderrors "errors"
	authdomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"log"
	"net/http"
)

type httpErr struct {
	status  int
	message string
}

// мапа внутренняя ошибка -> внешний ответ.
var errToHTTP = map[error]httpErr{
	authdomain.ErrUserAlreadyExists:  {status: http.StatusConflict, message: "user already exists"},
	authdomain.ErrInvalidCredentials: {status: http.StatusUnauthorized, message: "invalid credentials"},
	authdomain.ErrNoSession:          {status: http.StatusUnauthorized, message: "unauthorized"},
	authdomain.ErrInvalidToken:       {status: http.StatusUnauthorized, message: "unauthorized"},
	authdomain.ErrOther:              {status: http.StatusInternalServerError, message: "internal server error"},
}

type errorResponse struct {
	Error string `json:"error"`
}

// перекладывать статус и тело в writeError

func WriteServiceError(w http.ResponseWriter, err error) {
	var key error

	switch {
	case stderrors.Is(err, authdomain.ErrUserAlreadyExists):
		key = authdomain.ErrUserAlreadyExists
	case stderrors.Is(err, authdomain.ErrInvalidCredentials):
		key = authdomain.ErrInvalidCredentials
	case stderrors.Is(err, authdomain.ErrNoSession):
		key = authdomain.ErrNoSession
	case stderrors.Is(err, authdomain.ErrInvalidToken):
		key = authdomain.ErrInvalidToken
	default:
		key = authdomain.ErrOther
	}

	mapped := errToHTTP[key]

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(mapped.status)

	err = json.NewEncoder(w).Encode(errorResponse{Error: mapped.message})
	if err != nil {
		log.Printf("error marshaling error: %v\n", err)
	}
}
