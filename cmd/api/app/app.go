package app

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/config"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/server"
	authapp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth"
)

func Run() error {
	configPath := flag.String("config", "", "config file path")
	flag.Parse()

	cfg := &Config{}
	
	if err := config.LoadConfig(*configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config %w", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server started on %d", cfg.Server.Port)
	
	//создал роутер
	mux := http.NewServeMux()

	// ручки на signIn + signUp + refresh
	authService := authapp.NewService(cfg.Auth)
	authHandler := authapp.NewHandler(authService)
	authHandler.RegisterRoutes(mux)

	//запуск сервера (я думаю можно несколько запусков, если на мапах будут мьютексы)
	return server.RunServer(addr, mux, cfg.Server.Timeouts)
}
