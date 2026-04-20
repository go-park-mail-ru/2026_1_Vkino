package auth

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/config"
	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/delivery/grpc"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/repository/postgres"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/clock"
	jwtsvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/jwt"
	passwordsvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/password"
	authusecase "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/usecase"
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

	appLogger := baseLogger.WithField("component", "auth-service")

	options := corepostgres.BuildPostgresOptions(&cfg.Postgres)
	pgDB, err := corepostgres.New(cfg.Postgres, options...)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pgDB.Close()

	appLogger.Info("successfully connected to postgres")

	userRepo := postgresrepo.NewUserRepo(pgDB)
	sessionRepo := postgresrepo.NewSessionRepo(pgDB)

	clockService := clocksvc.New()
	passwordService := passwordsvc.New()
	jwtService := jwtsvc.New(jwtsvc.Config{
		Secret: cfg.Auth.JWTSecret,
		Issuer: cfg.Auth.Issuer,
	})

	authUC := authusecase.NewAuthUsecase(
		userRepo,
		sessionRepo,
		jwtService,
		passwordService,
		clockService,
		cfg.Auth,
	)

	lis, err := newListener(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := newGRPCServer(authUC)

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

func newAuthServer(u authusecase.Usecase) *deliverygrpc.Server {
	return deliverygrpc.NewServer(u)
}
