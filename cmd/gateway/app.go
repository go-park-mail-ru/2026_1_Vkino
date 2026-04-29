package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/routes"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	rootmw "github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	"google.golang.org/grpc"
)

func Run(configPath string) error {
	cfg := &Config{}
	if err := Load(configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", "api-gateway")
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = metrics.StartServer(runCtx, "api-gateway", cfg.Metrics, appLogger); err != nil {
		return fmt.Errorf("start metrics server: %w", err)
	}

	authConn, err := newGRPCConn(cfg.AuthGRPC)
	if err != nil {
		return fmt.Errorf("init auth grpc client: %w", err)
	}

	defer func() {
		_ = authConn.Close()
	}()

	userConn, err := newGRPCConn(cfg.UserGRPC)
	if err != nil {
		return fmt.Errorf("init user grpc client: %w", err)
	}

	defer func() {
		_ = userConn.Close()
	}()

	movieConn, err := newGRPCConn(cfg.MovieGRPC)
	if err != nil {
		return fmt.Errorf("init movie grpc client: %w", err)
	}

	defer func() {
		_ = movieConn.Close()
	}()

	server := httpserver.New(append(serverOptions(cfg, appLogger), routes.Register(
		cfg,
		authv1.NewAuthServiceClient(authConn),
		routes.NewUserClient(userConn, movieConn),
		moviev1.NewMovieServiceClient(movieConn),
	)...)...)

	appLogger.WithField("port", cfg.Server.Port).Info("starting api gateway")

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Run()
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-errCh:
		cancel()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server stopped with error: %w", err)
		}
	case sig := <-stopCh:
		cancel()
		appLogger.WithField("signal", sig.String()).Info("shutting down api gateway")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			return fmt.Errorf("shutdown api gateway: %w", shutdownErr)
		}

		err = <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server stopped with error: %w", err)
		}
	}

	return nil
}

func newGRPCConn(cfg ServiceGRPCConfig) (*grpc.ClientConn, error) {
	return grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.Address,
		RequestTimeout: cfg.RequestTimeout,
	})
}

func serverOptions(cfg *Config, log *logger.Logger) []httpserver.Option {
	return []httpserver.Option{
		httpserver.Port(cfg.Server.Port),
		httpserver.Timeout(cfg.Server.Timeouts),
		httpserver.WithMiddleware(rootmw.RequestIDMiddleware),
		httpserver.WithMiddleware(rootmw.CorsMiddleware(rootmw.CORSConfig{
			AllowedOrigins:   cfg.Server.CORS.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
			AllowCredentials: cfg.Server.CORS.AllowCredentials,
			MaxAge:           3600,
		})),
		httpserver.WithMiddleware(rootmw.LoggerMiddleware(log)),
		httpserver.WithMiddleware(rootmw.RecoveryMiddleware),
	}
}
