package user

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/config"
	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/delivery/grpc"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/repository/postgres"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/service/clock"
	userusecase "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/usecase"

	"google.golang.org/grpc"
)

func Run(configPath string) error {
	cfg := &config.Config{}
	if err := config.Load(configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", "user-service")

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

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, func(server *grpc.Server) {
		userv1.RegisterUserServiceServer(server, deliverygrpc.NewServer(userUC))
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
