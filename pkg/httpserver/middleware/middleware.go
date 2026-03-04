package middleware

import (
	"context"
	"net/http"
	"strings"

	authHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/delivery/http"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
)

type ctxKey string

const UserEmailKey ctxKey = "user_email"

type AuthMiddleware struct {
	usecase *usecase.AuthUsecase
}

func NewAuthMiddleware(u *usecase.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{usecase: u}
}

// Middleware валидирует access token и кладёт email пользователя в context.
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			authHttp.WriteError(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			authHttp.WriteError(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		if accessToken == "" {
			authHttp.WriteError(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		email, err := m.usecase.ValidateAccessToken(accessToken)
		if err != nil {
			authHttp.WriteError(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		ctx := context.WithValue(r.Context(), UserEmailKey, email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserEmailFromContext читает email из context в хэндлерах.
func UserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)

	return email, ok
}
