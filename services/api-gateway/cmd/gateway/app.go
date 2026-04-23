package main

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	rootmw "github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/movie/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
	routes "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/app/gateway/routes"
	authmw "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http/middleware"
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

	authConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.AuthGRPC.Address,
		RequestTimeout: cfg.AuthGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init auth grpc client: %w", err)
	}
	defer authConn.Close()

	userConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.UserGRPC.Address,
		RequestTimeout: cfg.UserGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init user grpc client: %w", err)
	}
	defer userConn.Close()

	movieConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.MovieGRPC.Address,
		RequestTimeout: cfg.MovieGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init movie grpc client: %w", err)
	}
	defer movieConn.Close()

	authClient := authv1.NewAuthServiceClient(authConn)
	userClient := userv1.NewUserServiceClient(userConn)
	movieClient := moviev1.NewMovieServiceClient(movieConn)

	authMiddleware := authmw.NewAuthMiddleware(authClient, grpcx.ClientConfig{
		Address:        cfg.AuthGRPC.Address,
		RequestTimeout: cfg.AuthGRPC.RequestTimeout,
	})

	opts := []httpserver.Option{
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
		httpserver.WithMiddleware(rootmw.LoggerMiddleware(appLogger)),
		httpserver.WithMiddleware(rootmw.RecoveryMiddleware),
	}

	opts = append(opts, routes.Register(
		cfg,
		authClient,
		userClient,
		movieClient,
		authMiddleware.Middleware,
	)...)

	server := httpserver.New(opts...)

	appLogger.WithField("port", cfg.Server.Port).Info("starting api gateway")

	return server.Run()
}
