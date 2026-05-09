package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

func (s *service) GetOverview(ctx context.Context, userID int64) (domain.OverviewResponse, error) {
	if userID <= 0 {
		return domain.OverviewResponse{}, domain.ErrInvalidUserID
	}

	return domain.OverviewResponse{}, domain.ErrNotImplemented
}

func (s *service) GetRoom(ctx context.Context, userID, roomID int64) (domain.RoomResponse, error) {
	if userID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidUserID
	}

	if roomID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidRoomID
	}

	return domain.RoomResponse{}, domain.ErrNotImplemented
}

func (s *service) CreateRoom(
	ctx context.Context,
	userID int64,
	req domain.CreateRoomRequest,
) (domain.RoomResponse, error) {
	if userID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidUserID
	}

	if req.Name == "" {
		return domain.RoomResponse{}, domain.ErrInvalidRoomName
	}

	if req.Visibility == "" {
		return domain.RoomResponse{}, domain.ErrInvalidVisibility
	}

	return domain.RoomResponse{}, domain.ErrNotImplemented
}

func (s *service) JoinRoom(
	ctx context.Context,
	userID int64,
	req domain.JoinRoomRequest,
) (domain.RoomResponse, error) {
	if userID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 && req.InviteLink == "" {
		return domain.RoomResponse{}, domain.ErrInvalidInviteLink
	}

	return domain.RoomResponse{}, domain.ErrNotImplemented
}

func (s *service) DeleteRoom(ctx context.Context, userID, roomID int64) (domain.DeleteRoomResponse, error) {
	if userID <= 0 {
		return domain.DeleteRoomResponse{}, domain.ErrInvalidUserID
	}

	if roomID <= 0 {
		return domain.DeleteRoomResponse{}, domain.ErrInvalidRoomID
	}

	return domain.DeleteRoomResponse{}, domain.ErrNotImplemented
}

func (s *service) SubscribeRoom(
	ctx context.Context,
	userID int64,
	req domain.SubscribeRoomRequest,
) (<-chan domain.RoomEvent, func(), error) {
	if userID <= 0 {
		return nil, nil, domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return nil, nil, domain.ErrInvalidRoomID
	}

	return nil, nil, domain.ErrNotImplemented
}
