package middleware

import (
	"context"
	"net/http"
	"runtime/debug"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func(ctx context.Context) {
			if err := recover(); err != nil {
				logger.FromContext(ctx).
					WithField("panic", err).
					WithField("stack", string(debug.Stack())).
					Error("panic recovered")
				http.Error(w, "500 - Internal Server Error", http.StatusInternalServerError)
			}
		}(r.Context())

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
