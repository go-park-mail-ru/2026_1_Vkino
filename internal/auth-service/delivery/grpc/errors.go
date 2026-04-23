package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/auth-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/auth-service/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/grpcx"
	"google.golang.org/grpc/codes"
)

var authGRPCErrorMapper = grpcx.New(
	[]error{
		domain.ErrUserAlreadyExists,
		postgresrepo.ErrUserAlreadyExists,

		domain.ErrInvalidCredentials,
		domain.ErrPasswordMismatch,
		domain.ErrNoSession,
		domain.ErrInvalidToken,

		domain.ErrUserNotFound,
		postgresrepo.ErrUserNotFound,

		domain.ErrInternal,
	},
	map[error]grpcx.ErrResponse{
		domain.ErrUserAlreadyExists:       {Code: codes.AlreadyExists, Message: "user already exists"},
		postgresrepo.ErrUserAlreadyExists: {Code: codes.AlreadyExists, Message: "user already exists"},

		domain.ErrInvalidCredentials: {Code: codes.Unauthenticated, Message: "unauthorized"},
		domain.ErrPasswordMismatch:   {Code: codes.Unauthenticated, Message: "unauthorized"},
		domain.ErrNoSession:          {Code: codes.Unauthenticated, Message: "unauthorized"},
		domain.ErrInvalidToken:       {Code: codes.Unauthenticated, Message: "unauthorized"},

		domain.ErrUserNotFound:       {Code: codes.NotFound, Message: "user not found"},
		postgresrepo.ErrUserNotFound: {Code: codes.NotFound, Message: "user not found"},

		domain.ErrInternal: {Code: codes.Internal, Message: "internal server error"},
	},
	codes.Internal,
	"internal server error",
)

func mapError(err error) error {
	return authGRPCErrorMapper.Map(err)
}
