package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository"
)

func NewSupportUsecase(
	supportRepo repository.SupportRepo,
	userRepo repository.UserRepo,
) *supportUsecase {
	return &supportUsecase{
		supportRepo: supportRepo,
		userRepo:    userRepo,
		broker:      newTicketBroker(),
	}
}
