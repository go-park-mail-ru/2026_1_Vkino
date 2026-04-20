package gateway

import (
	"context"
	"fmt"
	"net/http"

	rootmw "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"

	authgrpc "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/client/authgrpc"
	moviegrpc "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/client/moviegrpc"
	usergrpc "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/client/usergrpc"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/config"
	deliveryhttp "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http"
	authmw "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http/middleware"
	authusecase "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/usecase/auth"
	movieusecase "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/usecase/movie"
	userusecase "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/usecase/user"
)

type AuthMiddleware interface {
	Middleware(next http.Handler) http.Handler
}

func Run(configPath string) error {
	cfg := &config.Config{}
	if err := config.Load(configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", "api-gateway")

	authClient, err := authgrpc.New(context.Background(), authgrpc.Config{
		Address:        cfg.AuthGRPC.Address,
		RequestTimeout: cfg.AuthGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init auth grpc client: %w", err)
	}
	defer authClient.Close()

	userClient, err := usergrpc.New(context.Background(), usergrpc.Config{
		Address:        cfg.UserGRPC.Address,
		RequestTimeout: cfg.UserGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init user grpc client: %w", err)
	}
	defer userClient.Close()

	movieClient, err := moviegrpc.New(context.Background(), moviegrpc.Config{
		Address:        cfg.MovieGRPC.Address,
		RequestTimeout: cfg.MovieGRPC.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("init movie grpc client: %w", err)
	}
	defer movieClient.Close()

	authFacade := authusecase.NewFacade(authClient, cfg.UserAuth)
	userFacade := userusecase.NewFacade(userClient)
	movieFacade := movieusecase.NewFacade(movieClient)

	authHandler := deliveryhttp.NewAuthHandler(authFacade)
	userHandler := deliveryhttp.NewUserHandler(userFacade)
	movieHandler := deliveryhttp.NewMovieHandler(movieFacade)
	healthHandler := deliveryhttp.NewHealthHandler()

	authMiddleware := authmw.NewAuthMiddleware(authClient)

	opts := []httpserver.Option{
		httpserver.Port(cfg.Server.Port),
		httpserver.Timeout(cfg.Server.Timeouts),

		httpserver.WithMiddleware(rootmw.CorsMiddleware(rootmw.CORSConfig{
			AllowedOrigins:   cfg.Server.CORS.AllowedOrigins,
			AllowCredentials: cfg.Server.CORS.AllowCredentials,
			//MaxAge:           cfg.Server.CORS.MaxAge,
		})),
		httpserver.WithMiddleware(rootmw.LoggerMiddleware(appLogger)),
		httpserver.WithMiddleware(rootmw.RecoveryMiddleware),
	}

	opts = append(opts, registerRoutes(
		healthHandler,
		authHandler,
		userHandler,
		movieHandler,
		authMiddleware,
	)...)

	server := httpserver.New(opts...)

	appLogger.WithField("port", cfg.Server.Port).Info("starting api gateway")

	return server.Run()
}
