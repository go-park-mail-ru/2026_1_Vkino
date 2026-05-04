package main

import (
	"context"
	"fmt"

	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/delivery/grpc"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/repository/postgres"
	movieusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serverrunner"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"

	"google.golang.org/grpc"
)

const (
	serviceName   = "movie-service"
	componentName = "movie"
)

//nolint:gocyclo,cyclop // Service wiring intentionally stays explicit in the entrypoint.
func Run(configPath string) error {
	cfg := Config{}
	if err := Load(configPath, &cfg); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", componentName)

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = metrics.StartServer(runCtx, serviceName, cfg.Metrics, appLogger); err != nil {
		return fmt.Errorf("start metrics server: %w", err)
	}

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

	if err = posterStore.EnsureBucket(context.Background(), cfg.S3.Region); err != nil {
		return fmt.Errorf("ensure poster bucket: %w", err)
	}

	cardStore, err := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketCards))
	if err != nil {
		return fmt.Errorf("init card storage: %w", err)
	}

	if err = cardStore.EnsureBucket(context.Background(), cfg.S3.Region); err != nil {
		return fmt.Errorf("ensure card bucket: %w", err)
	}

	actorStore, err := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketActors))
	if err != nil {
		return fmt.Errorf("init actor storage: %w", err)
	}

	if err = actorStore.EnsureBucket(context.Background(), cfg.S3.Region); err != nil {
		return fmt.Errorf("ensure actor bucket: %w", err)
	}

	videoStore, err := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketVideos))
	if err != nil {
		return fmt.Errorf("init video storage: %w", err)
	}

	if err = videoStore.EnsureBucket(context.Background(), cfg.S3.Region); err != nil {
		return fmt.Errorf("ensure video bucket: %w", err)
	}

	movieRepo := postgresrepo.NewMovieRepo(pgDB)

	movieUC := movieusecase.NewMovieUsecase(
		movieRepo,
		posterStore,
		cardStore,
		actorStore,
		videoStore,
	)

	authConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.AuthGRPC.Address,
		RequestTimeout: cfg.AuthGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init auth grpc client: %w", err)
	}

	defer func() {
		_ = authConn.Close()
	}()

	authClient := authv1.NewAuthServiceClient(authConn)

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, serviceName, func(server *grpc.Server) {
		moviev1.RegisterMovieServiceServer(server, deliverygrpc.NewServer(movieUC, authClient))
	})

	appLogger.WithField("port", cfg.GRPC.Port).Info("starting grpc server")

	return serverrunner.RunGRPC(
		runCtx,
		appLogger,
		serviceName,
		func() error {
			return grpcServer.Serve(lis)
		},
		grpcServer.GracefulStop,
		grpcServer.Stop,
	)
}
