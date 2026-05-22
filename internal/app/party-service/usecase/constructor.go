package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/repository"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/subscription"
)

func New(
	partyRepo repository.PartyRepo,
	eventBroker repository.RoomEventBroker,
	subscriptionReader subscription.StateReader,
) Usecase {
	return &service{
		partyRepo:          partyRepo,
		eventBroker:        eventBroker,
		subscriptionReader: subscriptionReader,
	}
}
