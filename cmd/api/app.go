package main

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/cmd/api/app"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"

	userHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/delivery/http"
	userUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"

	movieHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/delivery/http"
	movieUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

func Run(configPath *string) error {
	cfg := &app.Config{}

	if err := app.LoadConfig(*configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config %w", err)
	}

	baseLogger, err := logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	appLogger := baseLogger.WithField("component", "api")

	options := postgres.BuildPostgresOptions(&cfg.Postgres)
	pgDB, err := postgres.New(cfg.Postgres, options...)

	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	appLogger.Info("successfully connected to postgres")

	defer pgDB.Close()

	userRepo := postgres.NewUserRepo(pgDB)
	sessionRepo := postgres.NewSessionRepo(pgDB)
	movieRepo := postgres.NewMovieRepo(pgDB)

	s3CommonConfig := cfg.S3.Config()

	actorStorage, err := storagepkg.NewS3Storage(context.Background(), s3CommonConfig.WithBucket(cfg.S3.BucketActors))
	if err != nil {
		return fmt.Errorf("init actor storage: %w", err)
	}

	posterStorage, err := storagepkg.NewS3Storage(context.Background(), s3CommonConfig.WithBucket(cfg.S3.BucketPosters))
	if err != nil {
		return fmt.Errorf("init poster storage: %w", err)
	}

	cardStorage, err := storagepkg.NewS3Storage(context.Background(), s3CommonConfig.WithBucket(cfg.S3.BucketCards))
	if err != nil {
		return fmt.Errorf("init card storage: %w", err)
	}

	videoStorage, err := storagepkg.NewS3Storage(context.Background(), s3CommonConfig.WithBucket(cfg.S3.BucketVideos))
	if err != nil {
		return fmt.Errorf("init video storage: %w", err)
	}

	avatarStorage, err := storagepkg.NewS3Storage(context.Background(), s3CommonConfig.WithBucket(cfg.S3.BucketAvatars))
	if err != nil {
		return fmt.Errorf("init avatar storage: %w", err)
	}

	userUsecase := userUsecase.NewAuthUsecaseWithStorage(userRepo, sessionRepo, avatarStorage, cfg.User)
	movieUsecase := movieUsecase.NewMovieUsecase(movieRepo, actorStorage, posterStorage, cardStorage, videoStorage)

	userHandler := userHttp.NewHandler(userUsecase)
	movieHandler := movieHttp.NewHandler(movieUsecase)

	authMiddleware := middleware.NewAuthMiddleware(userUsecase)

	corsMiddleware := middleware.CorsMiddleware(cfg.CORS)

	server := httpserver.New(
		httpserver.Port(cfg.Server.Port),
		httpserver.Timeout(cfg.Server.Timeouts),

		httpserver.WithMiddleware(corsMiddleware),
		httpserver.WithMiddleware(middleware.LoggerMiddleware(appLogger)),
		httpserver.WithMiddleware(middleware.RecoveryMiddleware),

		httpserver.WithRoute("POST /user/sign-up", userHandler.SignUp),
		httpserver.WithRoute("POST /user/sign-in", userHandler.SignIn),
		httpserver.WithRoute("POST /user/refresh", userHandler.Refresh),
		httpserver.WithRoute("GET /movie/selection/all", movieHandler.GetAllSelections),
		httpserver.WithRoute("GET /movie/selection/{selection}", movieHandler.GetSelectionByTitle),
		httpserver.WithRoute("GET /movie/search", movieHandler.Search),
		httpserver.WithRoute("GET /movie/{id}", movieHandler.GetMovieByID),
		httpserver.WithRoute("GET /movie/actor/{id}", movieHandler.GetActorByID),
		httpserver.WithRoute("GET /episode/{id}/playback", movieHandler.GetEpisodePlayback),

		httpserver.WithMiddlewareRoute("GET /user/me", userHandler.Me, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("PUT /user/profile", userHandler.UpdateProfile, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /user/change-password", userHandler.ChangePassword, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /user/logout", userHandler.LogOut, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("GET /episode/{id}/progress", movieHandler.GetEpisodeProgress,
			authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("PUT /episode/{id}/progress", movieHandler.SaveEpisodeProgress,
			authMiddleware.Middleware),
	)

	appLogger.WithField("port", cfg.Server.Port).Info("starting http server")

	return server.Run()
}
