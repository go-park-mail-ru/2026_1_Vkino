package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/auth-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
	usecase usecase.Usecase
}

func NewServer(u usecase.Usecase) *Server {
	return &Server{usecase: u}
}
