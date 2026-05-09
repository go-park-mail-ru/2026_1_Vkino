package main

import (
	"context"
	"fmt"

	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/delivery/grpc"
	partyusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serverrunner"

	"google.golang.org/grpc"
)

const (
	serviceName   = "party-service"
	componentName = "party"
)

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

	movieConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.MovieGRPC.Address,
		RequestTimeout: cfg.MovieGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init movie grpc client: %w", err)
	}
	defer func() {
		_ = movieConn.Close()
	}()

	userConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.UserGRPC.Address,
		RequestTimeout: cfg.UserGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init user grpc client: %w", err)
	}
	defer func() {
		_ = userConn.Close()
	}()

	partyUC := partyusecase.New()

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, serviceName, func(server *grpc.Server) {
		partyv1.RegisterPartyServiceServer(server, deliverygrpc.NewServer(
			partyUC,
			authv1.NewAuthServiceClient(authConn),
			moviev1.NewMovieServiceClient(movieConn),
			userv1.NewUserServiceClient(userConn),
		))
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
