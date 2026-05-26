package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

func (s *Server) authorize(ctx context.Context) (context.Context, authctx.Context, error) {
	authCtx, err := authctx.ValidateIncomingContext(ctx, s.authClient)
	if err != nil {
		return ctx, authctx.Context{}, err
	}

	ctx = authctx.WithContext(ctx, authCtx)

	return ctx, authCtx, nil
}
