package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/respond"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/auth/v1"
)

type ctxKey string

const AuthCtxKey ctxKey = "gateway_auth"

var ErrAuthContextMissing = errors.New("gateway auth context missing")

type AuthContext struct {
	UserID int64
	Email  string
}

type AuthMiddleware struct {
	authClient authv1.AuthServiceClient
	timeout    grpcx.ClientConfig
}

func NewAuthMiddleware(authClient authv1.AuthServiceClient, cfg grpcx.ClientConfig) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
		timeout:    cfg,
	}
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		if accessToken == "" {
			httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx, cancel := grpcx.WithTimeout(r.Context(), m.timeout.RequestTimeout)
		defer cancel()

		resp, err := m.authClient.Validate(ctx, &authv1.ValidateRequest{
			AccessToken: accessToken,
		})
		if err != nil {
			respond.Error(w, err)
			return
		}

		authCtx := AuthContext{
			UserID: resp.GetUserId(),
			Email:  resp.GetEmail(),
		}

		ctx = context.WithValue(r.Context(), AuthCtxKey, authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthFromContext(ctx context.Context) (AuthContext, error) {
	authCtx, ok := ctx.Value(AuthCtxKey).(AuthContext)
	if !ok {
		return AuthContext{}, ErrAuthContextMissing
	}

	return authCtx, nil
}
