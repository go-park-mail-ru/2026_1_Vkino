package main

import (
	"context"
	"fmt"

	deliverygrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/delivery/grpc"
	usergrpc "github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository/grpc"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository/postgres"
	yookassaclient "github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository/yookassa"
	paymentusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	paymentv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/payment/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serverrunner"

	"google.golang.org/grpc"
)

const (
	serviceName   = "payment-service"
	componentName = "payment"
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

	pgDB, err := corepostgres.New(cfg.Postgres, corepostgres.BuildPostgresOptions(&cfg.Postgres)...)
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
	defer func() { _ = authConn.Close() }()

	userConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.UserGRPC.Address,
		RequestTimeout: cfg.UserGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init user grpc client: %w", err)
	}
	defer func() { _ = userConn.Close() }()

	paymentRepo := postgresrepo.NewPaymentRepo(pgDB)
	yookassaClient := yookassaclient.NewClient(yookassaclient.Config{
		APIURL:    cfg.YooKassa.APIURL,
		ShopID:    cfg.YooKassa.ShopID,
		SecretKey: cfg.YooKassa.SecretKey,
		ReturnURL: cfg.YooKassa.ReturnURL,
		Capture:   cfg.YooKassa.Capture,
		Timeout:   cfg.YooKassa.Timeout,
	})
	activator := usergrpc.NewUserSubscriptionActivator(userv1.NewUserServiceClient(userConn))

	paymentUC := paymentusecase.New(
		paymentRepo,
		yookassaClient,
		activator,
		cfg.YooKassa.ReturnURL,
		cfg.YooKassa.Capture,
	)

	lis, err := grpcx.Listen(cfg.GRPC.Port)
	if err != nil {
		return err
	}

	grpcServer := grpcx.NewServer(appLogger, serviceName, func(server *grpc.Server) {
		authClient := authv1.NewAuthServiceClient(authConn)
		paymentv1.RegisterPaymentServiceServer(server, deliverygrpc.NewServer(paymentUC, authClient))
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
