package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/clock"
	jwtsvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/jwt"
	passwordsvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/password"
)

func NewAuthUsecase(
	userRepo repository.UserRepo,
	sessionRepo repository.SessionRepo,
	jwtService jwtsvc.Service,
	passwordService passwordsvc.Service,
	clockService clocksvc.Service,
	cfg Config,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		jwtService:      jwtService,
		passwordService: passwordService,
		clockService:    clockService,
		cfg:             cfg,
	}
}

func (u *AuthUsecase) GetConfig() Config {
	return u.cfg
}
