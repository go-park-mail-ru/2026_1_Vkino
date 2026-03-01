package server

import (
	"time"
	"net/http"
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

func RunServer(addr string, handler http.Handler, t TimeoutsConfig) error {
    s := http.Server{
        Addr:    addr,
        Handler: handler,

        ReadHeaderTimeout: t.ReadHeader,
        ReadTimeout:       t.Read,
        WriteTimeout:      t.Write,
        IdleTimeout:       t.Idle,
    }
    return s.ListenAndServe()
}