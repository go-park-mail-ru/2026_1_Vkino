package middleware

import (
	"net/http"
	"slices"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := range slices.Backward(middlewares) {
		h = middlewares[i](h)
	}

	return h
}
