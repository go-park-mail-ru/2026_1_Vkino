package metrics

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serverrunner"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
}

var errMetricsContextRequired = errors.New("metrics context is required")

func StartServer(ctx context.Context, service string, cfg Config, log *logger.Logger) error {
	addr, ok := metricsAddress(cfg)
	if !ok {
		return nil
	}

	if ctx == nil {
		return errMetricsContextRequired
	}

	Register()
	SetServiceInfo(service)

	baseLog := log
	if baseLog == nil {
		baseLog = logger.FromContext(ctx)
	}

	var listenConfig net.ListenConfig

	lis, err := listenConfig.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen metrics on %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: serverrunner.DefaultShutdownTimeout,
	}

	metricsLog := baseLog.
		WithField("component", "metrics").
		WithField("service", labelValue(service)).
		WithField("address", addr)

	startMetricsShutdown(ctx, server, metricsLog)
	startMetricsServe(server, lis, metricsLog)

	return nil
}

func metricsAddress(cfg Config) (string, bool) {
	if !cfg.Enabled {
		return "", false
	}

	addr := strings.TrimSpace(cfg.Address)
	if addr == "" {
		return "", false
	}

	return addr, true
}

func startMetricsShutdown(ctx context.Context, server *http.Server, log *logger.Logger) {
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			serverrunner.DefaultShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithField("error", err.Error()).Error("metrics shutdown failed")
		}
	}()
}

func startMetricsServe(server *http.Server, lis net.Listener, log *logger.Logger) {
	go func() {
		log.Info("starting metrics server")

		if err := server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithField("error", err.Error()).Error("metrics server stopped with error")
		}
	}()
}
