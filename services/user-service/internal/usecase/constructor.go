package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/service/clock"
)

func NewUserUsecase(
	userRepo repository.UserRepo,
	avatarStore storage.FileStorage,
	clockService clocksvc.Service,
) *UserUsecase {
	return &UserUsecase{
		userRepo:     userRepo,
		avatarStore:  avatarStore,
		clockService: clockService,
	}
}
