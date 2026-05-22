package main

import (
	"context"
	"fmt"

	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/delivery/grpc"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
	userusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serverrunner"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/clock"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"

	"google.golang.org/grpc"
)

const (
	serviceName   = "user-service"
	componentName = "user"
)

func Run(configPath string) error {
	cfg, appLogger, runCtx, err := bootstrapUserApp(configPath)
	if err != nil {
		return err
	}
	defer runCtx.cancel()

	pgDB, err := openUserPostgres(cfg.Postgres, appLogger)
	if err != nil {
		return err
	}
	defer pgDB.Close()

	avatarStore, supportFileStore, err := initUserStores(cfg)
	if err != nil {
		return err
	}

	userRepo := postgresrepo.NewUserRepo(pgDB)
	supportRepo := postgresrepo.NewSupportRepo(pgDB)
	clockService := clocksvc.New()

	userUC := userusecase.NewUserUsecase(userRepo, avatarStore, clockService)
	supportUC := userusecase.NewSupportUsecase(supportRepo, userRepo, supportFileStore, clockService)

	authConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.AuthGRPC.Address,
		RequestTimeout: cfg.AuthGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init auth grpc client: %w", err)
	}

	defer func() { _ = authConn.Close() }()

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, serviceName, func(server *grpc.Server) {
		authClient := authv1.NewAuthServiceClient(authConn)
		userv1.RegisterUserServiceServer(server, deliverygrpc.NewServer(userUC, authClient))
		supportv1.RegisterSupportServiceServer(server, deliverygrpc.NewSupportServer(supportUC, authClient))
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

func bootstrapUserApp(configPath string) (Config, *logger.Logger, appRunContext, error) {
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

func openUserPostgres(cfg corepostgres.Config, log *logger.Logger) (*corepostgres.Client, error) {
	pgDB, err := corepostgres.New(cfg, corepostgres.BuildPostgresOptions(&cfg)...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	log.Info("successfully connected to postgres")

	return pgDB, nil
}

func initUserStores(cfg Config) (*storage.S3Storage, storage.FileStorage, error) {
	avatarStore, err := newEnsuredStore(cfg.S3.Config().WithBucket(cfg.S3.BucketAvatars), cfg.S3.Region, "avatar")
	if err != nil {
		return nil, nil, err
	}

	supportStore, _, err := optionalSupportStore(cfg)
	if err != nil {
		return nil, nil, err
	}

	return avatarStore, supportStore, nil
}

func optionalSupportStore(cfg Config) (storage.FileStorage, bool, error) {
	if cfg.S3.BucketSupport == "" {
		return nil, false, nil
	}

	store, err := newEnsuredStore(cfg.S3.Config().WithBucket(cfg.S3.BucketSupport), cfg.S3.Region, "support file")
	if err != nil {
		return nil, false, err
	}

	return store, true, nil
}

type appRunContext struct {
	ctx    context.Context
	cancel context.CancelFunc
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
