package main

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/api-gateway/routes"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	rootmw "github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
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

	authConn, err := newGRPCConn(cfg.AuthGRPC)
	if err != nil {
		return fmt.Errorf("init auth grpc client: %w", err)
	}
	defer authConn.Close()

	userConn, err := newGRPCConn(cfg.UserGRPC)
	if err != nil {
		return fmt.Errorf("init user grpc client: %w", err)
	}
	defer userConn.Close()

	movieConn, err := newGRPCConn(cfg.MovieGRPC)
	if err != nil {
		return fmt.Errorf("init movie grpc client: %w", err)
	}
	defer movieConn.Close()

	server := httpserver.New(append(serverOptions(cfg, appLogger), routes.Register(
		cfg,
		authv1.NewAuthServiceClient(authConn),
		userv1.NewUserServiceClient(userConn),
		moviev1.NewMovieServiceClient(movieConn),
	)...)...)

	appLogger.WithField("port", cfg.Server.Port).Info("starting api gateway")

	return server.Run()
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
