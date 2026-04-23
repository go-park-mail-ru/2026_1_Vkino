package grpcx

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ClientConfig struct {
	Address        string
	RequestTimeout time.Duration
}

func Dial(ctx context.Context, cfg ClientConfig) (*grpc.ClientConn, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("grpc address is empty")
	}

	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", cfg.Address, err)
	}

	return conn, nil
}

func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)

	if id, ok := requestid.FromContext(ctx); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, requestid.MetadataKey, id)
	}

	return ctx, cancel
}
