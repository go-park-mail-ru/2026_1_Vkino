package main

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/routes"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	rootmw "github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
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

	var supportFileStore storage.FileStorage
	if cfg.S3.BucketSupport != "" {
		s3Store, storageErr := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketSupport))
		if storageErr != nil {
			return fmt.Errorf("init support file storage: %w", storageErr)
		}

		if storageErr = s3Store.EnsureBucket(context.Background(), cfg.S3.Region); storageErr != nil {
			return fmt.Errorf("ensure support file bucket: %w", storageErr)
		}

		supportFileStore = s3Store
	}

	server := httpserver.New(append(serverOptions(cfg, appLogger), routes.Register(
		cfg,
		authv1.NewAuthServiceClient(authConn),
		routes.NewUserClient(userConn),
		supportFileStore,
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
