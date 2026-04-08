package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2026_1_VKino/cmd/api/app"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"

	authHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/delivery/http"
	authUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"

	movieHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/delivery/http"
	movieUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
	profileHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/delivery/http"
	profileUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

func Run(configPath *string) error {
	cfg := &app.Config{}

	if err := app.LoadConfig(*configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config %w", err)
	}

	log.Printf("Server started on %d", cfg.Server.Port)

	options := postgres.BuildPostgresOptions(&cfg.Postgres)
	pgDB, err := postgres.New(cfg.Postgres, options...)

	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	log.Println("successfully connected to postgres")

	defer pgDB.Close()

	userRepo := postgres.NewUserRepo(pgDB)
	sessionRepo := postgres.NewSessionRepo(pgDB)
	movieRepo := postgres.NewMovieRepo(pgDB)

	s3Storage, err := storagepkg.NewS3Storage(context.Background(), storagepkg.Config{
		InternalEndpoint: cfg.S3.InternalEndpoint,
		PublicEndpoint:   cfg.S3.PublicEndpoint,
		Region:           cfg.S3.Region,
		AccessKeyID:      cfg.S3.AccessKeyID,
		SecretAccessKey:  cfg.S3.SecretAccessKey,
		Bucket:           cfg.S3.BucketImages,
		UseSSL:           cfg.S3.UseSSL,
		UsePathStyle:     cfg.S3.UsePathStyle,
		PresignTTL:       cfg.S3.PresignTTL,
	})
	if err != nil {
		return fmt.Errorf("init image storage: %w", err)
	}

	authUsecase := authUsecase.NewAuthUsecase(userRepo, sessionRepo, cfg.Auth)
	movieUsecase := movieUsecase.NewMovieUsecase(movieRepo, s3Storage)
	profileUsecase := profileUsecase.NewProfileUsecase(userRepo)

	authHandler := authHttp.NewHandler(authUsecase)
	movieHandler := movieHttp.NewHandler(movieUsecase)
	profileHandler := profileHttp.NewHandler(profileUsecase)

	authMiddleware := middleware.NewAuthMiddleware(authUsecase)

	corsMiddleware := middleware.CorsMiddleware(cfg.CORS)

	server := httpserver.New(
		httpserver.Port(cfg.Server.Port),
		httpserver.Timeout(cfg.Server.Timeouts),

		httpserver.WithMiddleware(corsMiddleware),
		httpserver.WithMiddleware(middleware.RecoveryMiddleware),

		httpserver.WithRoute("POST /auth/sign-up", authHandler.SignUp),
		httpserver.WithRoute("POST /auth/sign-in", authHandler.SignIn),
		httpserver.WithRoute("POST /auth/refresh", authHandler.Refresh),
		httpserver.WithRoute("GET /movie/selection/all", movieHandler.GetAllSelections),
		httpserver.WithRoute("GET /movie/selection/{selection}", movieHandler.GetSelectionByTitle),
		httpserver.WithRoute("GET /movie/{id}", movieHandler.GetMovieByID),
		httpserver.WithRoute("GET /movie/actor/{id}", movieHandler.GetActorByID),

		httpserver.WithMiddlewareRoute("GET /auth/me", profileHandler.GetProfile, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("GET /profile/me", profileHandler.GetProfile, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /auth/logout", authHandler.LogOut, authMiddleware.Middleware),

		// httpserver.WithRoute("GET /movie/{moviename}", movieHandler.GetMovieById) -- страница для проверки зарега
	)

	return server.Run()
}
