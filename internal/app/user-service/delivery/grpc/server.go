package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
)

type Server struct {
	userv1.UnimplementedUserServiceServer

	usecase    usecase.Usecase
	authClient authv1.AuthServiceClient
}

func NewServer(u usecase.Usecase, authClient authv1.AuthServiceClient) *Server {
	return &Server{
		usecase:    u,
		authClient: authClient,
	}
}
