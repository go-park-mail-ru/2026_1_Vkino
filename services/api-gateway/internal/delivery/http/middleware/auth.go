package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	authgrpc "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/client/authgrpc"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"

	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

type ctxKey string

const AuthCtxKey ctxKey = "gateway_auth"

var ErrAuthContextMissing = errors.New("gateway auth context missing")

type AuthMiddleware struct {
	authClient authgrpc.Client
}

func NewAuthMiddleware(authClient authgrpc.Client) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient}
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

		authCtx, err := m.authClient.Validate(r.Context(), accessToken)
		if err != nil {
			httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), AuthCtxKey, authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthFromContext(ctx context.Context) (dto.AuthContext, error) {
	authCtx, ok := ctx.Value(AuthCtxKey).(dto.AuthContext)
	if !ok {
		return dto.AuthContext{}, ErrAuthContextMissing
	}

	return authCtx, nil
}
