package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"google.golang.org/grpc/codes"
)

var userGRPCErrorMapper = grpcx.New(
	[]error{
		domain.ErrInvalidToken,

		domain.ErrUserNotFound,
		postgresrepo.ErrUserNotFound,

		domain.ErrInvalidSearchQuery,
		domain.ErrInvalidMovieID,
		domain.ErrInvalidBirthdate,
		domain.ErrInvalidAvatar,
		storage.ErrInvalidFileType,
		storage.ErrFileTooLarge,

		domain.ErrAlreadyFriends,
		domain.ErrFriendNotFound,
		domain.ErrSelfFriendship,

		domain.ErrInternal,
	},
	map[error]grpcx.ErrResponse{
		domain.ErrInvalidToken: {Code: codes.Unauthenticated, Message: "unauthorized"},

		domain.ErrUserNotFound:       {Code: codes.NotFound, Message: "user not found"},
		postgresrepo.ErrUserNotFound: {Code: codes.NotFound, Message: "user not found"},

		domain.ErrInvalidSearchQuery: {Code: codes.InvalidArgument, Message: "invalid search query"},
		domain.ErrInvalidMovieID:     {Code: codes.InvalidArgument, Message: "invalid movie id"},
		domain.ErrInvalidBirthdate:   {Code: codes.InvalidArgument, Message: "invalid birthdate"},
		domain.ErrInvalidAvatar:      {Code: codes.InvalidArgument, Message: "invalid avatar"},
		storage.ErrInvalidFileType:   {Code: codes.InvalidArgument, Message: "unsupported file extension"},
		storage.ErrFileTooLarge:      {Code: codes.InvalidArgument, Message: "file size exceeds the limit"},

		domain.ErrAlreadyFriends: {Code: codes.AlreadyExists, Message: "already friends"},
		domain.ErrFriendNotFound: {Code: codes.NotFound, Message: "friend not found"},
		domain.ErrSelfFriendship: {Code: codes.FailedPrecondition, Message: "self friendship is forbidden"},

		domain.ErrInternal: {Code: codes.Internal, Message: "internal server error"},
	},
	codes.Internal,
	"internal server error",
)

func mapError(err error) error {
	return userGRPCErrorMapper.Map(err)
}
