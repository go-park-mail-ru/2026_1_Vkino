package main

import (
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2026_1_VKino/cmd/api/app"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"

	authHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/delivery/http"
	authDomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	authUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"

	movieHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/delivery/http"
	movieDomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	movieUsecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/inmemory"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

func Run(configPath *string) error {
	cfg := &app.Config{}

	if err := auth.LoadConfig(*configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config %w", err)
	}

	log.Printf("Server started on %d", cfg.Server.Port)

	db := inmemory.NewDB([]inmemory.Named{
		&authDomain.User{},
		&authDomain.TokenPair{},
		&movieDomain.SelectionResponse{},
	})

	userRepo := inmemory.NewUserRepo(db)
	sessionRepo := inmemory.NewSessionRepo(db)
	movieRepo := inmemory.NewMovieRepo(db)

	authUsecase := authUsecase.NewAuthUsecase(userRepo, sessionRepo, cfg.Auth)
	movieUsecase := movieUsecase.NewMovieUsecase(movieRepo)

	authHandler := authHttp.NewHandler(authUsecase)
	movieHandler := movieHttp.NewHandler(movieUsecase)

	authMiddleware := middleware.NewAuthMiddleware(authUsecase)

	server := httpserver.New(
		httpserver.Port(cfg.Server.Port),
		httpserver.Timeout(cfg.Server.Timeouts),

		httpserver.WithMiddleware(middleware.CorsMiddleware),
		httpserver.WithMiddleware(middleware.RecoveryMiddleware),

		httpserver.WithRoute("POST /auth/sign-up", authHandler.SignUp),
		httpserver.WithRoute("POST /auth/sign-in", authHandler.SignIn),
		httpserver.WithRoute("POST /auth/refresh", authHandler.Refresh),
		httpserver.WithRoute("GET /movie/selection/all", movieHandler.GetAllSelections),
		httpserver.WithRoute("GET /movie/selection/{selection}", movieHandler.GetSelectionByTitle),

		httpserver.WithMiddlewareRoute("GET /auth/me", authHandler.Me, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /auth/logout", authHandler.LogOut, authMiddleware.Middleware),

		// httpserver.WithRoute("GET /movie/{moviename}", movieHandler.GetMovieById) -- страница для проверки зарега
	)

	return server.Run()
}
