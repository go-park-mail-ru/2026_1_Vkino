//nolint:interfacebloat // Central usecase contract intentionally groups room and realtime operations.
package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/repository"
)

type Usecase interface {
	GetOverview(ctx context.Context, userID int64) (domain.OverviewResponse, error)
	GetRoom(ctx context.Context, userID, roomID int64) (domain.RoomResponse, error)
	CreateRoom(ctx context.Context, userID int64, req domain.CreateRoomRequest) (domain.RoomResponse, error)
	JoinRoom(ctx context.Context, userID int64, req domain.JoinRoomRequest) (domain.RoomResponse, error)
	DeleteRoom(ctx context.Context, userID, roomID int64) (domain.DeleteRoomResponse, error)
	SubscribeRoom(
		ctx context.Context,
		userID int64,
		req domain.SubscribeRoomRequest,
	) (<-chan domain.RoomEvent, func(), error)
}

type service struct {
	partyRepo   repository.PartyRepo
	eventBroker repository.RoomEventBroker
}
