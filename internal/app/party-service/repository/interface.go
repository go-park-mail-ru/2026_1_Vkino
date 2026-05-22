package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

type PartyOverviewRepo interface {
	GetOverview(ctx context.Context, userID int64) (domain.OverviewResponse, error)
}

type PartyRoomRepo interface {
	GetRoomByID(ctx context.Context, roomID int64) (*domain.Room, error)
	CreateRoom(ctx context.Context, hostUserID int64, req domain.CreateRoomRequest) (*domain.Room, error)
	InviteMember(ctx context.Context, roomID, userID int64) error
	AddMember(ctx context.Context, roomID, userID int64) (*domain.Room, error)
	ActivateMember(ctx context.Context, roomID, userID int64) error
	DeleteRoom(ctx context.Context, roomID int64) error
	GetInvite(ctx context.Context, inviteLink string) (*domain.Invite, error)
}

type PartyPlaybackRepo interface {
	SavePlaybackState(ctx context.Context, roomID int64, state domain.PlaybackState) error
	TouchRoom(ctx context.Context, roomID int64) error
}

type PartyMessagingRepo interface {
	SaveMessage(ctx context.Context, message domain.RoomMessage) (*domain.RoomMessage, error)
}

type PartyPollRepo interface {
	SavePoll(ctx context.Context, poll domain.Poll) (*domain.Poll, error)
	SaveVote(ctx context.Context, vote domain.PollVote) error
}

type PartyRepo interface {
	PartyOverviewRepo
	PartyRoomRepo
	PartyPlaybackRepo
	PartyMessagingRepo
	PartyPollRepo
}

type RoomEventBroker interface {
	Publish(ctx context.Context, event domain.RoomEvent) error
	Subscribe(ctx context.Context, roomID, userID int64) (<-chan domain.RoomEvent, func(), error)
	ActiveUsers(roomID int64) int32
	IsUserActive(roomID, userID int64) bool
}
