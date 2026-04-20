package grpc

import (
	"errors"

	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/repository/postgres"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, domain.ErrInvalidMovieID),
		errors.Is(err, domain.ErrInvalidActorID),
		errors.Is(err, domain.ErrInvalidSelectionTitle):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrMovieNotFound),
		errors.Is(err, postgresrepo.ErrMovieNotFound):
		return status.Error(codes.NotFound, "movie not found")

	case errors.Is(err, domain.ErrActorNotFound),
		errors.Is(err, postgresrepo.ErrActorNotFound):
		return status.Error(codes.NotFound, "actor not found")

	case errors.Is(err, domain.ErrSelectionNotFound),
		errors.Is(err, postgresrepo.ErrSelectionNotFound):
		return status.Error(codes.NotFound, "selection not found")

	case errors.Is(err, domain.ErrInternal):
		return status.Error(codes.Internal, "internal server error")

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
