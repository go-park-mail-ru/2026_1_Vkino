package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	authHttp "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

type ctxKey string

const AuthCtxKey ctxKey = "auth"

type AuthMiddleware struct {
	usecase usecase.Usecase
}

func NewAuthMiddleware(u *usecase.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{usecase: u}
}

// Middleware валидирует access token и кладёт email и id пользователя в context.
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			authHttp.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			authHttp.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		if accessToken == "" {
			authHttp.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		auth, err := m.usecase.ValidateAccessToken(accessToken)
		if err != nil {
			authHttp.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		ctx := context.WithValue(r.Context(), AuthCtxKey, auth)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthFromContext читает AuthContext из context в хэндлерах.
func AuthFromContext(ctx context.Context) (usecase.AuthContext, error) {
	auth, ok := ctx.Value(AuthCtxKey).(usecase.AuthContext)
	if !ok {
		return usecase.AuthContext{}, ErrMidlware
	}

	return auth, nil
}
