package httpserver

import (
	"context"
	"net/http"
	"time"

	rootmw "github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/middleware"
)

type TimeoutsConfig struct {
	ReadHeader time.Duration `mapstructure:"read_header"`
	Read       time.Duration `mapstructure:"read"`
	Write      time.Duration `mapstructure:"write"`
	Idle       time.Duration `mapstructure:"idle"`
}

type CORSConfig struct {
	AllowedOrigins   []string      `mapstructure:"allowed_origins"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

type Config struct {
	Port     int            `mapstructure:"port"`
	Timeouts TimeoutsConfig `mapstructure:"timeouts"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

type Middleware = rootmw.Middleware

const defaultReadHeaderTimeout = 5 * time.Second

type Server struct {
	server      *http.Server
	mux         *http.ServeMux
	middlewares []Middleware
}

func (s *Server) Handler() http.Handler {
	return s.server.Handler
}

func New(opts ...Option) *Server {
	mux := http.NewServeMux()
	s := &Server{
		mux:         mux,
		middlewares: []Middleware{},
		server: &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.server.Handler = s.applyMiddlewares(mux)

	return s
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) applyMiddlewares(h http.Handler) http.Handler {
	return rootmw.Chain(h, s.middlewares...)
}
