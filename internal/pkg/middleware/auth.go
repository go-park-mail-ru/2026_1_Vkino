package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	authHttp "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

type ctxKey string

const UserEmailKey ctxKey = "user_email"

type AuthMiddleware struct {
	usecase usecase.Usecase
}

func NewAuthMiddleware(u *usecase.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{usecase: u}
}

// Middleware валидирует access token и кладёт email пользователя в context.
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

		email, err := m.usecase.ValidateAccessToken(accessToken)
		if err != nil {
			authHttp.ErrResponse(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		ctx := context.WithValue(r.Context(), UserEmailKey, email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserEmailFromContext читает email из context в хэндлерах.
func UserEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(UserEmailKey).(string)
	if !ok {
		return "", ErrMidlware
	}
	return email, nil
}
