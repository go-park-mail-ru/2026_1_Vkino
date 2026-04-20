package grpc

import (
	"errors"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/repository/postgres"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "unauthorized")

	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, postgresrepo.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	case errors.Is(err, domain.ErrInvalidSearchQuery),
		errors.Is(err, domain.ErrInvalidMovieID),
		errors.Is(err, domain.ErrInvalidBirthdate),
		errors.Is(err, domain.ErrInvalidAvatar),
		errors.Is(err, storage.ErrInvalidFileType):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrAlreadyFriends):
		return status.Error(codes.AlreadyExists, "already friends")

	case errors.Is(err, domain.ErrFriendNotFound):
		return status.Error(codes.NotFound, "friend not found")

	case errors.Is(err, domain.ErrSelfFriendship):
		return status.Error(codes.FailedPrecondition, "self friendship is forbidden")

	case errors.Is(err, domain.ErrInternal):
		return status.Error(codes.Internal, "internal server error")

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
