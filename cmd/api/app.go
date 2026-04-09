package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2026_1_VKino/cmd/api/app"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"

	userHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/delivery/http"
	userUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"

	movieHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/delivery/http"
	movieUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
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

	userUsecase := userUsecase.NewAuthUsecaseWithStorage(userRepo, sessionRepo, s3Storage, cfg.User)
	movieUsecase := movieUsecase.NewMovieUsecase(movieRepo, s3Storage)

	userHandler := userHttp.NewHandler(userUsecase)
	movieHandler := movieHttp.NewHandler(movieUsecase)

	authMiddleware := middleware.NewAuthMiddleware(userUsecase)

	corsMiddleware := middleware.CorsMiddleware(cfg.CORS)

	server := httpserver.New(
		httpserver.Port(cfg.Server.Port),
		httpserver.Timeout(cfg.Server.Timeouts),

		httpserver.WithMiddleware(corsMiddleware),
		httpserver.WithMiddleware(middleware.RecoveryMiddleware),

		httpserver.WithRoute("POST /user/sign-up", userHandler.SignUp),
		httpserver.WithRoute("POST /user/sign-in", userHandler.SignIn),
		httpserver.WithRoute("POST /user/refresh", userHandler.Refresh),
		httpserver.WithRoute("GET /movie/selection/all", movieHandler.GetAllSelections),
		httpserver.WithRoute("GET /movie/selection/{selection}", movieHandler.GetSelectionByTitle),
		httpserver.WithRoute("GET /movie/{id}", movieHandler.GetMovieByID),
		httpserver.WithRoute("GET /movie/actor/{id}", movieHandler.GetActorByID),

		httpserver.WithMiddlewareRoute("GET /user/me", userHandler.Me, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("PUT /user/profile", userHandler.UpdateProfile, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /user/change-password", userHandler.ChangePassword, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /user/logout", userHandler.LogOut, authMiddleware.Middleware),

		// httpserver.WithRoute("GET /movie/{moviename}", movieHandler.GetMovieById) -- страница для проверки зарега
	)

	return server.Run()
}
