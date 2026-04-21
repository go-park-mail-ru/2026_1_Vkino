package interceptor

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryRecovery(log *logger.Logger) grpc.UnaryServerInterceptor {
	baseLog := log
	if baseLog == nil {
		baseLog = logger.FromContext(context.Background())
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				requestLog := baseLog.
					WithField("transport", "grpc").
					WithField("grpc_method", info.FullMethod)

				if id, ok := requestid.FromContext(ctx); ok {
					requestLog = requestLog.WithField("request_id", id)
				}

				requestLog.
					WithField("panic", fmt.Sprintf("%v", rec)).
					Error("grpc panic recovered")

				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}