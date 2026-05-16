//nolint:interfacebloat // Central repository contracts intentionally group party persistence and realtime primitives.
package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

type PartyRepo interface {
	GetOverview(ctx context.Context, userID int64) (domain.OverviewResponse, error)
	GetRoomByID(ctx context.Context, roomID int64) (*domain.Room, error)
	CreateRoom(ctx context.Context, hostUserID int64, req domain.CreateRoomRequest) (*domain.Room, error)
	InviteMember(ctx context.Context, roomID, userID int64) error
	AddMember(ctx context.Context, roomID, userID int64) (*domain.Room, error)
	ActivateMember(ctx context.Context, roomID, userID int64) error
	DeleteRoom(ctx context.Context, roomID int64) error
	GetInvite(ctx context.Context, inviteLink string) (*domain.Invite, error)
	SavePlaybackState(ctx context.Context, roomID int64, state domain.PlaybackState) error
	SaveMessage(ctx context.Context, message domain.RoomMessage) (*domain.RoomMessage, error)
	SavePoll(ctx context.Context, poll domain.Poll) (*domain.Poll, error)
	SaveVote(ctx context.Context, vote domain.PollVote) error
	TouchRoom(ctx context.Context, roomID int64) error
}

type RoomEventBroker interface {
	Publish(ctx context.Context, event domain.RoomEvent) error
	Subscribe(ctx context.Context, roomID int64) (<-chan domain.RoomEvent, func(), error)
}
