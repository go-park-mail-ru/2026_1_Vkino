package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/user-service/delivery/grpc"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/user-service/repository/postgres"
	userusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/user-service/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/clock"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/auth/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"

	"google.golang.org/grpc"
)

func Run(configPath string) error {
	cfg := Config{}
	if err := Load(configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", "user")

	options := corepostgres.BuildPostgresOptions(&cfg.Postgres)
	pgDB, err := corepostgres.New(cfg.Postgres, options...)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pgDB.Close()

	appLogger.Info("successfully connected to postgres")

	avatarStore, err := storage.NewS3Storage(context.Background(), cfg.S3.Config().WithBucket(cfg.S3.BucketAvatars))
	if err != nil {
		return fmt.Errorf("init avatar storage: %w", err)
	}

	userRepo := postgresrepo.NewUserRepo(pgDB)
	clockService := clocksvc.New()

	userUC := userusecase.NewUserUsecase(userRepo, avatarStore, clockService)

	authConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.AuthGRPC.Address,
		RequestTimeout: cfg.AuthGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init auth grpc client: %w", err)
	}
	defer authConn.Close()

	authClient := authv1.NewAuthServiceClient(authConn)

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, func(server *grpc.Server) {
		userv1.RegisterUserServiceServer(server, deliverygrpc.NewServer(userUC, authClient))
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
