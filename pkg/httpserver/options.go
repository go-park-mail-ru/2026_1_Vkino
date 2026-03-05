package httpserver

import (
	"fmt"
	"net/http"
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
