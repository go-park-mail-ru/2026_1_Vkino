package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/clock"
	jwtsvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/jwt"
	passwordsvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/password"
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
