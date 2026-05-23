package main

import (
	"context"
	"fmt"
	"time"

	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/delivery/grpc"
	memoryrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/repository/memory"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/repository/postgres"
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

	authConn, movieConn, userConn, closeConns, err := openPartyConns(cfg)
	if err != nil {
		return err
	}
	defer closeConns()

	partyRepo := postgresrepo.NewPartyRepo(pgDB)
	eventBroker := memoryrepo.NewRoomEventBroker()
	partyUC := partyusecase.New(
		partyRepo,
		eventBroker,
		partyusecase.NewSubscriptionReader(userv1.NewUserServiceClient(userConn)),
	)

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

func openPartyConns(
	cfg Config,
) (authConn, movieConn, userConn *grpc.ClientConn, closeFn func(), err error) {
	authConn, err = dialNamedGRPCConn("auth", cfg.AuthGRPC.Address, cfg.AuthGRPC.RequestTimeout)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	movieConn, err = dialNamedGRPCConn("movie", cfg.MovieGRPC.Address, cfg.MovieGRPC.RequestTimeout)
	if err != nil {
		_ = authConn.Close()

		return nil, nil, nil, nil, err
	}

	userConn, err = dialNamedGRPCConn("user", cfg.UserGRPC.Address, cfg.UserGRPC.RequestTimeout)
	if err != nil {
		_ = movieConn.Close()
		_ = authConn.Close()

		return nil, nil, nil, nil, err
	}

	return authConn, movieConn, userConn, func() {
		_ = userConn.Close()
		_ = movieConn.Close()
		_ = authConn.Close()
	}, nil
}

func dialNamedGRPCConn(name, address string, timeout time.Duration) (*grpc.ClientConn, error) {
	conn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        address,
		RequestTimeout: timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("init %s grpc client: %w", name, err)
	}

	return conn, nil
}
