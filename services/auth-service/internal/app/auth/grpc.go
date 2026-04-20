package auth

import (
	"fmt"
	"net"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/auth/v1"
	authusecase "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/usecase"

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

func newGRPCServer(u authusecase.Usecase) *grpc.Server {
	server := grpc.NewServer()
	authv1.RegisterAuthServiceServer(server, newAuthServer(u))
	reflection.Register(server)

	return server
}