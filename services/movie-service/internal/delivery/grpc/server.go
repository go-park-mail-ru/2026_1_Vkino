package grpc

import (
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/usecase"
)

type Server struct {
	moviev1.UnimplementedMovieServiceServer
	usecase usecase.Usecase
}

func NewServer(u usecase.Usecase) *Server {
	return &Server{usecase: u}
}
