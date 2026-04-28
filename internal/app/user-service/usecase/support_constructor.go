package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/clock"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

func NewSupportUsecase(
	supportRepo repository.SupportRepo,
	userRepo repository.UserRepo,
	supportFileStore storage.FileStorage,
	clockService clocksvc.Service,
) *supportUsecase {
	return &supportUsecase{
		supportRepo:      supportRepo,
		userRepo:         userRepo,
		supportFileStore: supportFileStore,
		clockService:     clockService,
		broker:           newTicketBroker(),
	}
}
