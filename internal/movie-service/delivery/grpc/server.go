package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/movie-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
)

type Server struct {
	moviev1.UnimplementedMovieServiceServer
	usecase    usecase.Usecase
	authClient authv1.AuthServiceClient
}

func NewServer(u usecase.Usecase, authClient authv1.AuthServiceClient) *Server {
	return &Server{
		usecase:    u,
		authClient: authClient,
	}
}
