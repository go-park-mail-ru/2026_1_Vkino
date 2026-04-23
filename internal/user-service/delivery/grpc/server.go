package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/user-service/usecase"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	usecase usecase.Usecase
}

func NewServer(u usecase.Usecase) *Server {
	return &Server{usecase: u}
}
