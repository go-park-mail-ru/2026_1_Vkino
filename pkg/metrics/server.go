package metrics

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
}

func StartServer(ctx context.Context, service string, cfg Config, log *logger.Logger) error {
	if !cfg.Enabled {
		return nil
	}

	addr := strings.TrimSpace(cfg.Address)
	if addr == "" {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	Register()
	SetServiceInfo(service)

	baseLog := log
	if baseLog == nil {
		baseLog = logger.FromContext(ctx)
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen metrics on %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	metricsLog := baseLog.
		WithField("component", "metrics").
		WithField("service", labelValue(service, "unknown")).
		WithField("address", addr)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			metricsLog.WithField("error", err.Error()).Error("metrics shutdown failed")
		}
	}()

	go func() {
		metricsLog.Info("starting metrics server")

		if err := server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			metricsLog.WithField("error", err.Error()).Error("metrics server stopped with error")
		}
	}()

	return nil
}
