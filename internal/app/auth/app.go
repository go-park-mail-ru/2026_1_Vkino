package auth

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	authHttp "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/delivery/http"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/repository/mapDB"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

// заглушки для mapDB.
type userModel struct{}

func (m userModel) Name() string { return "users" }

type sessionModel struct{}

func (m sessionModel) Name() string { return "sessions" }

func Run() error {
	configPath := flag.String("config", "", "config file path")

	flag.Parse()

	cfg := &Config{}

	if err := LoadConfig(*configPath, cfg); err != nil {
		return fmt.Errorf("unable to load config %w", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server started on %d", cfg.Server.Port)

	// создал роутер
	mux := http.NewServeMux()

	db := mapDB.NewDB([]mapDB.Named{userModel{}, sessionModel{}})

	userRepo := mapDB.NewUserRepo(db)
	sessionRepo := mapDB.NewSessionRepo(db)

	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, &cfg.Auth)

	authHandler := authHttp.NewHandler(authUsecase)
	authHandler.RegisterRoutes(mux)

	// запуск сервера (я думаю можно несколько запусков, если на мапах будут мьютексы)
	return httpserver.RunServer(addr, mux, cfg.Server.Timeouts)
}
