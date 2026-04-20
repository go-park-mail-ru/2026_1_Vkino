package grpc

import (
	"errors"

	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/repository/postgres"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, domain.ErrUserAlreadyExists),
		errors.Is(err, postgresrepo.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")

	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrPasswordMismatch),
		errors.Is(err, domain.ErrNoSession),
		errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "unauthorized")

	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, postgresrepo.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	case errors.Is(err, domain.ErrInternal):
		return status.Error(codes.Internal, "internal server error")

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}