package grpc

import (
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/auth/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/usecase"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
	usecase usecase.Usecase
}

func NewServer(u usecase.Usecase) *Server {
	return &Server{usecase: u}
}