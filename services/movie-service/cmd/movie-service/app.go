package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/config"
	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/delivery/grpc"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/repository/postgres"
	movieusecase "github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/usecase"

	"google.golang.org/grpc"
)

func Run(configPath string) error {
	cfg := &config.Config{}
	if err := Load(configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", "movie-service")

	options := corepostgres.BuildPostgresOptions(&cfg.Postgres)
	pgDB, err := corepostgres.New(cfg.Postgres, options...)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pgDB.Close()

	appLogger.Info("successfully connected to postgres")

	posterStore, err := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketPosters))
	if err != nil {
		return fmt.Errorf("init poster storage: %w", err)
	}

	cardStore, err := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketCards))
	if err != nil {
		return fmt.Errorf("init card storage: %w", err)
	}

	actorStore, err := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketActors))
	if err != nil {
		return fmt.Errorf("init actor storage: %w", err)
	}

	videoStore, err := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketVideos))
	if err != nil {
		return fmt.Errorf("init video storage: %w", err)
	}

	movieRepo := postgresrepo.NewMovieRepo(pgDB)

	movieUC := movieusecase.NewMovieUsecase(
		movieRepo,
		posterStore,
		cardStore,
		actorStore,
		videoStore,
	)

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, func(server *grpc.Server) {
		moviev1.RegisterMovieServiceServer(server, deliverygrpc.NewServer(movieUC))
	})

	appLogger.WithField("port", cfg.GRPC.Port).Info("starting grpc server")

	errCh := make(chan error, 1)
	go func() {
		errCh <- grpcServer.Serve(lis)
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-errCh:
		if err != nil {
			return fmt.Errorf("grpc server stopped with error: %w", err)
		}
	case sig := <-stopCh:
		appLogger.WithField("signal", sig.String()).Info("shutting down grpc server")
		grpcServer.GracefulStop()
	}

	return nil
}