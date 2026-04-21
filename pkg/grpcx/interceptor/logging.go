package interceptor

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/requestid"
	"google.golang.org/grpc"
)

func UnaryLogging(log *logger.Logger) grpc.UnaryServerInterceptor {
	baseLog := log
	if baseLog == nil {
		baseLog = logger.FromContext(context.Background())
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		requestLog := baseLog.
			WithField("transport", "grpc").
			WithField("grpc_method", info.FullMethod)

		if id, ok := requestid.FromContext(ctx); ok {
			requestLog = requestLog.WithField("request_id", id)
		}

		ctx = logger.ContextWithLogger(ctx, requestLog)

		resp, err := handler(ctx, req)

		durationMs := time.Since(start).Milliseconds()

		if err != nil {
			requestLog.
				WithField("duration_ms", durationMs).
				WithField("error", err.Error()).
				Error("grpc request failed")
			return resp, err
		}

		requestLog.
			WithField("duration_ms", durationMs).
			Info("grpc request completed")

		return resp, nil
	}
}