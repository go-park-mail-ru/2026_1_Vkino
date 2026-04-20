package movie

import (
	"fmt"
	"net"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/movie/v1"
	movieusecase "github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func newListener(port int) (net.Listener, error) {
	addr := fmt.Sprintf(":%d", port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen grpc on %s: %w", addr, err)
	}

	return lis, nil
}

func newGRPCServer(u movieusecase.Usecase) *grpc.Server {
	server := grpc.NewServer()
	moviev1.RegisterMovieServiceServer(server, newMovieServer(u))
	reflection.Register(server)

	return server
}
