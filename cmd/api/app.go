package main

import (
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth"
	authHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/delivery/http"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/inmemory"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

// заглушки для inmemory.
type userModel struct{}

func (m userModel) Name() string { return "users" }

type sessionModel struct{}

func (m sessionModel) Name() string { return "sessions" }

func Run(configPath *string) error {

	cfg := &auth.Config{}

	if err := auth.LoadConfig(*configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config %w", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server started on %d", cfg.Server.Port)

	db := inmemory.NewDB([]inmemory.Named{userModel{}, sessionModel{}})

	userRepo := inmemory.NewUserRepo(db)
	sessionRepo := inmemory.NewSessionRepo(db)

	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, cfg.Auth)

	authHandler := authHttp.NewHandler(authUsecase)

	server := httpserver.New(
		httpserver.Addr(addr),
		httpserver.Timeout(cfg.Server.Timeouts),
		httpserver.WithRoute("/sign-up", authHandler.SignUp),
		httpserver.WithRoute("/sign-in", authHandler.SignIn),
		httpserver.WithRoute("/refresh", authHandler.Refresh),
	)

	return server.Run()
}
