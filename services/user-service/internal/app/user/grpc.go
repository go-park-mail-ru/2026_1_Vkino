package user

import (
	"fmt"
	"net"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
	userusecase "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/usecase"

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

func newGRPCServer(u userusecase.Usecase) *grpc.Server {
	server := grpc.NewServer()
	userv1.RegisterUserServiceServer(server, newUserServer(u))
	reflection.Register(server)

	return server
}
