package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(service string) grpc.UnaryServerInterceptor {
	Register()

	serviceLabel := labelValue(service)

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		startedAt := time.Now()

		resp, err := handler(ctx, req)

		code := status.Code(err).String()
		method := labelValue(info.FullMethod)

		GRPCRequestsTotal.WithLabelValues(serviceLabel, method, code).Inc()
		GRPCRequestDurationSeconds.WithLabelValues(serviceLabel, method, code).
			Observe(time.Since(startedAt).Seconds())

		if err != nil {
			GRPCRequestErrorsTotal.WithLabelValues(serviceLabel, method, code).Inc()
		}

		return resp, err
	}
}

func StreamServerInterceptor(service string) grpc.StreamServerInterceptor {
	Register()

	serviceLabel := labelValue(service)

	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		startedAt := time.Now()

		err := handler(srv, ss)

		code := status.Code(err).String()
		method := labelValue(info.FullMethod)

		GRPCStreamsTotal.WithLabelValues(serviceLabel, method, code).Inc()
		GRPCStreamDurationSeconds.WithLabelValues(serviceLabel, method, code).
			Observe(time.Since(startedAt).Seconds())

		if err != nil {
			GRPCStreamErrorsTotal.WithLabelValues(serviceLabel, method, code).Inc()
		}

		return err
	}
}
