package errors

import (
	stderrors "errors"
	"net/http"

	moviedomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	userdomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	inmemoryrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/inmemory"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
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

	inmemoryrepo.ErrSelectionNotFound: {status: http.StatusNotFound, message: "selection not found"},
	inmemoryrepo.ErrMovieNotFound:     {status: http.StatusNotFound, message: "movie not found"},
	inmemoryrepo.ErrActorNotFound:     {status: http.StatusNotFound, message: "actor not found"},

	postgresrepo.ErrSelectionNotFound: {status: http.StatusNotFound, message: "selection not found"},
	postgresrepo.ErrMovieNotFound:     {status: http.StatusNotFound, message: "movie not found"},
	postgresrepo.ErrActorNotFound:     {status: http.StatusNotFound, message: "actor not found"},
	postgresrepo.ErrEpisodeNotFound:   {status: http.StatusNotFound, message: "episode not found"},

	moviedomain.ErrInvalidMovieID:       {status: http.StatusBadRequest, message: "invalid movie id"},
	moviedomain.ErrInvalidActorID:       {status: http.StatusBadRequest, message: "invalid actor id"},
	moviedomain.ErrInvalidEpisodeID:     {status: http.StatusBadRequest, message: "invalid episode id"},
	moviedomain.ErrInvalidWatchProgress: {status: http.StatusBadRequest, message: "invalid watch progress"},

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
