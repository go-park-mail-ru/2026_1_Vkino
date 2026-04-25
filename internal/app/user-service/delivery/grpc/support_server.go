package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/grpcx"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"google.golang.org/grpc/codes"
)

type SupportServer struct {
	supportv1.UnimplementedSupportServiceServer
	usecase    usecase.SupportUsecase
	authClient authv1.AuthServiceClient
}

func NewSupportServer(u usecase.SupportUsecase, authClient authv1.AuthServiceClient) *SupportServer {
	return &SupportServer{
		usecase:    u,
		authClient: authClient,
	}
}

var supportGRPCErrorMapper = grpcx.New(
	[]error{
		domain.ErrInvalidToken,
		domain.ErrTicketNotFound,
		postgresrepo.ErrTicketNotFound,
		domain.ErrAccessDenied,
		domain.ErrInvalidTicketID,
		domain.ErrInternal,
	},
	map[error]grpcx.ErrResponse{
		domain.ErrInvalidToken:          {Code: codes.Unauthenticated, Message: "unauthorized"},
		domain.ErrTicketNotFound:         {Code: codes.NotFound, Message: "ticket not found"},
		postgresrepo.ErrTicketNotFound:   {Code: codes.NotFound, Message: "ticket not found"},
		domain.ErrAccessDenied:           {Code: codes.PermissionDenied, Message: "access denied"},
		domain.ErrInvalidTicketID:        {Code: codes.InvalidArgument, Message: "invalid ticket id"},
		domain.ErrInternal:               {Code: codes.Internal, Message: "internal server error"},
	},
	codes.Internal,
	"internal server error",
)

func mapSupportError(err error) error {
	return supportGRPCErrorMapper.Map(err)
}
