package main

import (
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2026_1_VKino/cmd/api/app"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth"
	authHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/delivery/http"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
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
		&domain.User{},
		&domain.TokenPair{},
	})

	userRepo := inmemory.NewUserRepo(db)
	sessionRepo := inmemory.NewSessionRepo(db)

	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, cfg.Auth)

	authHandler := authHttp.NewHandler(authUsecase)

	server := httpserver.New(
		httpserver.Port(cfg.Server.Port),
		httpserver.Timeout(cfg.Server.Timeouts),
		httpserver.WithRoute("/sign-up", authHandler.SignUp),
		httpserver.WithRoute("/sign-in", authHandler.SignIn),
		httpserver.WithRoute("/refresh", authHandler.Refresh),
	)

	return server.Run()
}
