package main

import (
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2026_1_VKino/cmd/api/app"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"

	authHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/delivery/http"
	authUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"

	movieHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/delivery/http"
	movieUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

func Run(configPath *string) error {
	cfg := &app.Config{}

	if err := app.LoadConfig(*configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config %w", err)
	}

	log.Printf("Server started on %d", cfg.Server.Port)

	dsn := cfg.Postgres.DSN()

	options := postgres.BuildPostgresOptions(&cfg.Postgres)

	pgDB, err := postgres.New(dsn, options...)

	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	log.Println("successfully connected to postgres")

	defer pgDB.Close()

	userRepo := postgres.NewUserRepo(pgDB)
	sessionRepo := postgres.NewSessionRepo(pgDB)
	movieRepo := postgres.NewMovieRepo(pgDB)

	authUsecase := authUsecase.NewAuthUsecase(userRepo, sessionRepo, cfg.Auth)
	movieUsecase := movieUsecase.NewMovieUsecase(movieRepo)

	authHandler := authHttp.NewHandler(authUsecase)
	movieHandler := movieHttp.NewHandler(movieUsecase)

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

		httpserver.WithMiddlewareRoute("GET /auth/me", authHandler.Me, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /auth/logout", authHandler.LogOut, authMiddleware.Middleware),

		// httpserver.WithRoute("GET /movie/{moviename}", movieHandler.GetMovieById) -- страница для проверки зарега
	)

	return server.Run()
}
