package usecase

import "github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/repository"

func New(partyRepo repository.PartyRepo, eventBroker repository.RoomEventBroker) Usecase {
	return &service{
		partyRepo:   partyRepo,
		eventBroker: eventBroker,
	}
}
