package grpcx

import (
	"context"
	"fmt"
	"net"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx/interceptor"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Listen(port int) (net.Listener, error) {
	addr := fmt.Sprintf(":%d", port)

	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen grpc on %s: %w", addr, err)
	}

	return lis, nil
}

func NewServer(
	log *logger.Logger,
	serviceName string,
	register func(*grpc.Server),
	opts ...grpc.ServerOption,
) *grpc.Server {
	baseOpts := make([]grpc.ServerOption, 0, 1+len(opts))
	baseOpts = append(baseOpts,
		grpc.MaxRecvMsgSize(defaultMaxMessageSize),
		grpc.MaxSendMsgSize(defaultMaxMessageSize),
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryRequestID(),
			interceptor.UnaryLogging(log),
			metrics.UnaryServerInterceptor(serviceName),
			interceptor.UnaryRecovery(log),
		),
		grpc.ChainStreamInterceptor(
			metrics.StreamServerInterceptor(serviceName),
		),
	)

	baseOpts = append(baseOpts, opts...)

	server := grpc.NewServer(baseOpts...)

	if register != nil {
		register(server)
	}

	reflection.Register(server)

	return server
}
