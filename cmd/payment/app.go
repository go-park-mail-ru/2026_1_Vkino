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
	cfg, appLogger, runCtx, err := bootstrapPaymentApp(configPath)
	if err != nil {
		return err
	}
	defer runCtx.cancel()

	pgDB, err := openPaymentPostgres(cfg.Postgres, appLogger)
	if err != nil {
		return err
	}
	defer pgDB.Close()

	authConn, userConn, err := initPaymentGRPCClients(cfg)
	if err != nil {
		return err
	}

	defer func() { _ = authConn.Close() }()
	defer func() { _ = userConn.Close() }()

	paymentUC := newPaymentUsecase(cfg, pgDB, userConn)

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

type paymentRunContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func bootstrapPaymentApp(configPath string) (Config, *logger.Logger, paymentRunContext, error) {
	cfg := Config{}
	if err := Load(configPath, &cfg); err != nil {
		return Config{}, nil, paymentRunContext{}, fmt.Errorf("unable to load config: %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return Config{}, nil, paymentRunContext{}, fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", componentName)
	ctx, cancel := context.WithCancel(context.Background())
	runCtx := paymentRunContext{ctx: ctx, cancel: cancel}

	if err := metrics.StartServer(runCtx.ctx, serviceName, cfg.Metrics, appLogger); err != nil {
		cancel()

		return Config{}, nil, paymentRunContext{}, fmt.Errorf("start metrics server: %w", err)
	}

	return cfg, appLogger, runCtx, nil
}

func openPaymentPostgres(cfg corepostgres.Config, log *logger.Logger) (*corepostgres.Client, error) {
	pgDB, err := corepostgres.New(cfg, corepostgres.BuildPostgresOptions(&cfg)...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	log.Info("successfully connected to postgres")

	return pgDB, nil
}

func initPaymentGRPCClients(cfg Config) (*grpc.ClientConn, *grpc.ClientConn, error) {
	authConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.AuthGRPC.Address,
		RequestTimeout: cfg.AuthGRPC.RequestTimeout,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("init auth grpc client: %w", err)
	}

	userConn, err := grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.UserGRPC.Address,
		RequestTimeout: cfg.UserGRPC.RequestTimeout,
	})
	if err != nil {
		_ = authConn.Close()

		return nil, nil, fmt.Errorf("init user grpc client: %w", err)
	}

	return authConn, userConn, nil
}

func newPaymentUsecase(cfg Config, pgDB *corepostgres.Client, userConn *grpc.ClientConn) *paymentusecase.Usecase {
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

	return paymentusecase.New(
		paymentRepo,
		yookassaClient,
		activator,
		cfg.YooKassa.ReturnURL,
		cfg.YooKassa.Capture,
	)
}
