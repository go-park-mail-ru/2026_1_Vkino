package main

import (
	"context"
	"fmt"

	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/delivery/grpc"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/repository/postgres"
	movieusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
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

func Run(configPath string) error {
	cfg, appLogger, runCtx, err := bootstrapApp(configPath)
	if err != nil {
		return err
	}
	defer runCtx.cancel()

	pgDB, err := openPostgres(cfg.Postgres, appLogger)
	if err != nil {
		return err
	}
	defer pgDB.Close()

	stores, err := initMovieStores(cfg)
	if err != nil {
		return err
	}

	authConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.AuthGRPC.Address,
		RequestTimeout: cfg.AuthGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init auth grpc client: %w", err)
	}

	defer func() { _ = authConn.Close() }()

	userConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.UserGRPC.Address,
		RequestTimeout: cfg.UserGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init user grpc client: %w", err)
	}

	defer func() { _ = userConn.Close() }()

	movieUC := movieusecase.NewMovieUsecase(
		postgresrepo.NewMovieRepo(pgDB),
		movieusecase.NewSubscriptionReader(userv1.NewUserServiceClient(userConn)),
		stores.poster,
		stores.card,
		stores.actor,
		stores.video,
	)

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, serviceName, func(server *grpc.Server) {
		moviev1.RegisterMovieServiceServer(server, deliverygrpc.NewServer(movieUC, authv1.NewAuthServiceClient(authConn)))
	})

	appLogger.WithField("port", cfg.GRPC.Port).Info("starting grpc server")

	return serverrunner.RunGRPC(
		runCtx.ctx,
		appLogger,
		serviceName,
		func() error {
			return grpcServer.Serve(lis)
		},
		grpcServer.GracefulStop,
		grpcServer.Stop,
	)
}

type appRunContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type movieStores struct {
	poster *storage.S3Storage
	card   *storage.S3Storage
	actor  *storage.S3Storage
	video  *storage.S3Storage
}

func bootstrapApp(configPath string) (Config, *logger.Logger, appRunContext, error) {
	cfg := Config{}
	if err := Load(configPath, &cfg); err != nil {
		return Config{}, nil, appRunContext{}, fmt.Errorf("unable to load config: %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return Config{}, nil, appRunContext{}, fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", componentName)
	ctx, cancel := context.WithCancel(context.Background())
	runCtx := appRunContext{ctx: ctx, cancel: cancel}

	if err := metrics.StartServer(runCtx.ctx, serviceName, cfg.Metrics, appLogger); err != nil {
		cancel()

		return Config{}, nil, appRunContext{}, fmt.Errorf("start metrics server: %w", err)
	}

	return cfg, appLogger, runCtx, nil
}

func openPostgres(cfg corepostgres.Config, log *logger.Logger) (*corepostgres.Client, error) {
	pgDB, err := corepostgres.New(cfg, corepostgres.BuildPostgresOptions(&cfg)...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	log.Info("successfully connected to postgres")

	return pgDB, nil
}

func initMovieStores(cfg Config) (movieStores, error) {
	poster, err := newEnsuredStore(cfg.S3.Config().WithBucket(cfg.S3.BucketPosters), cfg.S3.Region, "poster")
	if err != nil {
		return movieStores{}, err
	}

	card, err := newEnsuredStore(cfg.S3.Config().WithBucket(cfg.S3.BucketCards), cfg.S3.Region, "card")
	if err != nil {
		return movieStores{}, err
	}

	actor, err := newEnsuredStore(cfg.S3.Config().WithBucket(cfg.S3.BucketActors), cfg.S3.Region, "actor")
	if err != nil {
		return movieStores{}, err
	}

	video, err := newEnsuredStore(cfg.S3.Config().WithBucket(cfg.S3.BucketVideos), cfg.S3.Region, "video")
	if err != nil {
		return movieStores{}, err
	}

	return movieStores{poster: poster, card: card, actor: actor, video: video}, nil
}

func newEnsuredStore(cfg storage.Config, region, name string) (*storage.S3Storage, error) {
	store, err := storage.NewS3Storage(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("init %s storage: %w", name, err)
	}

	if err := store.EnsureBucket(context.Background(), region); err != nil {
		return nil, fmt.Errorf("ensure %s bucket: %w", name, err)
	}

	return store, nil
}
