package grpc

import (
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/usecase"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	usecase usecase.Usecase
}

func NewServer(u usecase.Usecase) *Server {
	return &Server{usecase: u}
}
