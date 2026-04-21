package middleware

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/requestid"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestid.Normalize(r.Header.Get(requestid.HeaderName))

		w.Header().Set(requestid.HeaderName, id)

		ctx := requestid.ContextWithID(r.Context(), id)

		requestLog := logger.FromContext(ctx).WithField("request_id", id)
		ctx = logger.ContextWithLogger(ctx, requestLog)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}