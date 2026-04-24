package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/clock"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
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
