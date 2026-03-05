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

type Config struct {
	Port     int            `mapstructure:"port"`
	Timeouts TimeoutsConfig `mapstructure:"timeouts"`
}

type Server struct {
	server *http.Server
	mux    *http.ServeMux
}

func New(opts ...Option) *Server {
	mux := http.NewServeMux()
	s := &Server{
		mux: mux,
		server: &http.Server{
			Handler: mux,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}
