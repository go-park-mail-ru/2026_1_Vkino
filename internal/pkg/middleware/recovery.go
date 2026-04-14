package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/logger"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.FromContext(r.Context()).
					WithField("panic", err).
					WithField("stack", string(debug.Stack())).
					Error("panic recovered")
				http.Error(w, "500 - Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
