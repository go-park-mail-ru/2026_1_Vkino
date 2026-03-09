package httpserver

import (
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
)

type Option func(*Server)

func Port(port int) Option {
	return func(s *Server) {
		s.server.Addr = fmt.Sprintf(":%d", port)
	}
}

func Timeout(t TimeoutsConfig) Option {
	return func(s *Server) {
		s.server.ReadHeaderTimeout = t.ReadHeader
		s.server.ReadTimeout = t.Read
		s.server.WriteTimeout = t.Write
		s.server.IdleTimeout = t.Idle
	}
}

func WithRoute(pattern string, handler http.HandlerFunc) Option {
	return func(s *Server) {
		s.mux.HandleFunc(pattern, handler)
	}
}

func WithMiddlewareRoute(pattern string, handler http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		wrappedHandler := middleware.Chain(
			http.HandlerFunc(handler),
			middlewares...,
		)

		s.mux.Handle(pattern, wrappedHandler)
	}
}

func WithMiddleware(mw Middleware) Option {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, mw)
	}
}
