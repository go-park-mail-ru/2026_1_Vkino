package authctx

import (
	"context"
	"errors"
	"strings"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const MetadataKey = "authorization"

var (
	ErrAuthContextMissing         = errors.New("auth context missing")
	ErrAuthorizationHeaderMissing = errors.New("authorization header missing")
	ErrInvalidAuthorizationHeader = errors.New("invalid authorization header")
)

type Context struct {
	UserID int64
	Email  string
}

type ctxKey string

const authCtxKey ctxKey = "auth_context"

func WithContext(ctx context.Context, authCtx Context) context.Context {
	return context.WithValue(ctx, authCtxKey, authCtx)
}

func FromContext(ctx context.Context) (Context, error) {
	authCtx, ok := ctx.Value(authCtxKey).(Context)
	if !ok {
		return Context{}, ErrAuthContextMissing
	}

	return authCtx, nil
}

func AppendOutgoing(ctx context.Context, authorization string) context.Context {
	authorization = strings.TrimSpace(authorization)
	if authorization == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, MetadataKey, authorization)
}

func AccessTokenFromIncomingContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrAuthorizationHeaderMissing
	}

	values := md.Get(MetadataKey)
	if len(values) == 0 {
		return "", ErrAuthorizationHeaderMissing
	}

	return ParseBearerToken(values[0])
}

func ParseBearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", ErrAuthorizationHeaderMissing
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", ErrInvalidAuthorizationHeader
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
	if token == "" {
		return "", ErrInvalidAuthorizationHeader
	}

	return token, nil
}

func ValidateIncomingContext(ctx context.Context, authClient authv1.AuthServiceClient) (Context, error) {
	accessToken, err := AccessTokenFromIncomingContext(ctx)
	if err != nil {
		return Context{}, status.Error(codes.Unauthenticated, "unauthorized")
	}

	resp, err := authClient.Validate(ctx, &authv1.ValidateRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return Context{}, err
	}

	authCtx := Context{
		UserID: resp.GetUserId(),
		Email:  resp.GetEmail(),
	}

	return authCtx, nil
}
