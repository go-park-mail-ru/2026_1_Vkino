package main

import (
	"context"
	"fmt"

	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/repository/postgres"
	authusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serverrunner"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/clock"
	jwtsvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/jwt"
	passwordsvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/password"

	"google.golang.org/grpc"
)

const (
	serviceName   = "auth-service"
	componentName = "auth"
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

	userRepo := postgres.NewUserRepo(pgDB)
	sessionRepo := postgres.NewSessionRepo(pgDB)

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

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, serviceName, func(server *grpc.Server) {
		authv1.RegisterAuthServiceServer(server, deliverygrpc.NewServer(authUC))
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
