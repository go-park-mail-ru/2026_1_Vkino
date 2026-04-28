package interceptor

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryRequestID() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		id := ""

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get(requestid.MetadataKey)
			if len(values) > 0 {
				id = values[0]
			}
		}

		id = requestid.Normalize(id)
		ctx = requestid.ContextWithID(ctx, id)

		_ = grpc.SetHeader(ctx, metadata.Pairs(requestid.MetadataKey, id))

		return handler(ctx, req)
	}
}

