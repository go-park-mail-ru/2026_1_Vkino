package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpjson"
)

type ctxKey string
const UserEmailKey ctxKey = "user_email"

// Middleware  валидирует access token и кладёт email пользователя в context.
func (s *Service) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			httpjson.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			httpjson.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		if accessToken == "" {
			httpjson.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		email, err := s.validateAccessToken(accessToken)
		if err != nil {
			httpjson.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), UserEmailKey, email)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}


// чтение email из context в хэндлерах.
func UserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}