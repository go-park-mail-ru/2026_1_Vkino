package main

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/routes"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	rootmw "github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serverrunner"
	"google.golang.org/grpc"
)

const (
	serviceName       = "api-gateway"
	corsMaxAgeSeconds = 3600
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

	appLogger := baseLogger.WithField("component", serviceName)

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = metrics.StartServer(runCtx, serviceName, cfg.Metrics, appLogger); err != nil {
		return fmt.Errorf("start metrics server: %w", err)
	}

	authConn, userConn, movieConn, partyConn, closeConns, err := openGatewayConns(cfg)
	if err != nil {
		return err
	}
	defer closeConns()

	server := httpserver.New(append(serverOptions(cfg, appLogger), routes.Register(
		cfg,
		authv1.NewAuthServiceClient(authConn),
		routes.NewUserClient(userConn, movieConn),
		moviev1.NewMovieServiceClient(movieConn),
		partyv1.NewPartyServiceClient(partyConn),
	)...)...)

	appLogger.WithField("port", cfg.Server.Port).Info("starting api gateway")

	return serverrunner.RunHTTP(
		runCtx,
		appLogger,
		serviceName,
		server.Run,
		server.Shutdown,
	)
}

func openGatewayConns(
	cfg *Config,
) (authConn, userConn, movieConn, partyConn *grpc.ClientConn, closeFn func(), err error) {
	authConn, err = openNamedGRPCConn("auth", cfg.AuthGRPC)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	userConn, err = openNamedGRPCConn("user", cfg.UserGRPC)
	if err != nil {
		_ = authConn.Close()

		return nil, nil, nil, nil, nil, err
	}

	movieConn, err = openNamedGRPCConn("movie", cfg.MovieGRPC)
	if err != nil {
		_ = userConn.Close()
		_ = authConn.Close()

		return nil, nil, nil, nil, nil, err
	}

	partyConn, err = openNamedGRPCConn("party", cfg.PartyGRPC)
	if err != nil {
		_ = movieConn.Close()
		_ = userConn.Close()
		_ = authConn.Close()

		return nil, nil, nil, nil, nil, err
	}

	return authConn, userConn, movieConn, partyConn, func() {
		_ = partyConn.Close()
		_ = movieConn.Close()
		_ = userConn.Close()
		_ = authConn.Close()
	}, nil
}

func openNamedGRPCConn(name string, cfg ServiceGRPCConfig) (*grpc.ClientConn, error) {
	conn, err := newGRPCConn(cfg)
	if err != nil {
		return nil, fmt.Errorf("init %s grpc client: %w", name, err)
	}

	return conn, nil
}

func newGRPCConn(cfg ServiceGRPCConfig) (*grpc.ClientConn, error) {
	return grpcx.Dial(context.Background(), grpcx.ClientConfig{
		Address:        cfg.Address,
		RequestTimeout: cfg.RequestTimeout,
	})
}

func serverOptions(cfg *Config, log *logger.Logger) []httpserver.Option {
	return []httpserver.Option{
		httpserver.Port(cfg.Server.Port),
		httpserver.Timeout(cfg.Server.Timeouts),
		httpserver.WithMiddleware(rootmw.RequestIDMiddleware),
		httpserver.WithMiddleware(rootmw.CorsMiddleware(rootmw.CORSConfig{
			AllowedOrigins:   cfg.Server.CORS.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
			AllowCredentials: cfg.Server.CORS.AllowCredentials,
			MaxAge:           corsMaxAgeSeconds,
		})),
		httpserver.WithMiddleware(rootmw.LoggerMiddleware(log)),
		httpserver.WithMiddleware(rootmw.RecoveryMiddleware),
	}
}
