package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/grpcx"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
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

var supportGRPCErrorMapper = newGRPCErrorMapper(
	[]error{
		domain.ErrInvalidToken,
		domain.ErrInvalidEmail,
		domain.ErrTicketNotFound,
		postgresrepo.ErrTicketNotFound,
		domain.ErrAccessDenied,
		domain.ErrInvalidTicketID,
		domain.ErrInvalidTicketPayload,
		domain.ErrInvalidMessage,
		domain.ErrInvalidSupportFilePayload,
		storage.ErrInvalidFileType,
		storage.ErrFileTooLarge,
		storage.ErrStorageUnavailable,
		domain.ErrInternal,
	},
	mergeGRPCErrorRules(commonGRPCErrorRules, map[error]grpcx.ErrResponse{
		domain.ErrInvalidEmail:              {Code: codes.InvalidArgument, Message: "invalid email"},
		domain.ErrTicketNotFound:            {Code: codes.NotFound, Message: "ticket not found"},
		postgresrepo.ErrTicketNotFound:      {Code: codes.NotFound, Message: "ticket not found"},
		domain.ErrAccessDenied:              {Code: codes.PermissionDenied, Message: "access denied"},
		domain.ErrInvalidTicketID:           {Code: codes.InvalidArgument, Message: "invalid ticket id"},
		domain.ErrInvalidTicketPayload:      {Code: codes.InvalidArgument, Message: "invalid ticket payload"},
		domain.ErrInvalidMessage:            {Code: codes.InvalidArgument, Message: "invalid message payload"},
		domain.ErrInvalidSupportFilePayload: {Code: codes.InvalidArgument, Message: "invalid support file payload"},
		storage.ErrStorageUnavailable:       {Code: codes.Unavailable, Message: "support file storage is unavailable"},
	}),
)

func mapSupportError(err error) error {
	return supportGRPCErrorMapper.Map(err)
}
