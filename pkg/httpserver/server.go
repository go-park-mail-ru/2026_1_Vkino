package httpserver

import (
	"net/http"
	"time"
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

type Middleware func(http.Handler) http.Handler

type Server struct {
	server      *http.Server
	mux         *http.ServeMux
	middlewares []Middleware
}

func New(opts ...Option) *Server {
	mux := http.NewServeMux()
	s := &Server{
		mux:         mux,
		middlewares: []Middleware{},
		server: &http.Server{
			Handler: mux,
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

func (s *Server) applyMiddlewares(h http.Handler) http.Handler {
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		h = s.middlewares[i](h)
	}

	return h
}
