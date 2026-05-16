package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	"github.com/stretchr/testify/require"
)

func TestGetRoomRequiresMembershipEvenForPublicRoom(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{})

	_, err := svc.GetRoom(context.Background(), 3, 5)

	require.True(t, errors.Is(err, domain.ErrAccessDenied))
}

func TestGetRoomHidesInviteForNonHost(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{})

	resp, err := svc.GetRoom(context.Background(), 2, 5)
	require.NoError(t, err)
	require.Empty(t, resp.Room.InviteLink)
}

func TestGetRoomInviteAllowsHostOnly(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{})

	resp, err := svc.GetRoomInvite(context.Background(), 1, 5)
	require.NoError(t, err)
	require.Equal(t, int64(5), resp.RoomID)
	require.Equal(t, "invite-123", resp.InviteLink)
}

func TestInviteFriendToRoomCreatesPendingMember(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{})

	resp, err := svc.InviteFriendToRoom(context.Background(), 1, domain.InviteFriendToRoomRequest{
		RoomID:        5,
		InvitedUserID: 9,
	})
	require.NoError(t, err)
	require.Equal(t, int64(5), resp.RoomID)
	require.Equal(t, int64(9), resp.InvitedUserID)
	require.Equal(t, "pending", resp.Status)
	require.Equal(t, int64(5), repo.inviteMemberRoomID)
	require.Equal(t, int64(9), repo.inviteMemberUserID)
}

func TestGetRoomActivatesPendingMember(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	repo.room.Members = append(repo.room.Members, domain.RoomMember{UserID: 9, Role: "member", Status: "pending"})
	svc := New(repo, &roomUsecaseBroker{})

	resp, err := svc.GetRoom(context.Background(), 9, 5)
	require.NoError(t, err)
	require.Equal(t, int64(5), repo.activateMemberRoomID)
	require.Equal(t, int64(9), repo.activateMemberUserID)
	require.Equal(t, "active", findMemberStatus(t, resp.Room.Members, 9))
}

func TestJoinRoomUsesInviteOnly(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{})

	_, err := svc.JoinRoom(context.Background(), 9, domain.JoinRoomRequest{})

	require.True(t, errors.Is(err, domain.ErrInvalidInviteLink))
}

func TestJoinRoomAddsMemberByInviteAndHidesInviteForNonHost(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{})

	resp, err := svc.JoinRoom(context.Background(), 9, domain.JoinRoomRequest{
		InviteLink: "https://example.com/watch-party/join/invite-123",
	})
	require.NoError(t, err)
	require.Equal(t, int64(5), repo.addMemberRoomID)
	require.Equal(t, int64(9), repo.addMemberUserID)
	require.Empty(t, resp.Room.InviteLink)
}

func TestGetOverviewHidesInviteOutsideHostRooms(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{})

	resp, err := svc.GetOverview(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "invite-123", resp.ActiveRooms[0].InviteLink)
	require.Empty(t, resp.FeaturedRooms[0].InviteLink)
	require.Empty(t, resp.MyRooms[0].InviteLink)
}

func TestSubscribeRoomRequiresMembership(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	broker := &roomUsecaseBroker{}
	svc := New(repo, broker)

	_, _, err := svc.SubscribeRoom(context.Background(), 3, domain.SubscribeRoomRequest{RoomID: 5})

	require.True(t, errors.Is(err, domain.ErrAccessDenied))
	require.Empty(t, broker.subscribedRoomIDs)
}

type roomUsecaseRepo struct {
	room                 *domain.Room
	overview             domain.OverviewResponse
	invite               *domain.Invite
	inviteMemberRoomID   int64
	inviteMemberUserID   int64
	addMemberRoomID      int64
	addMemberUserID      int64
	activateMemberRoomID int64
	activateMemberUserID int64
}

func newRoomUsecaseRepo() *roomUsecaseRepo {
	room := &domain.Room{
		ID:         5,
		Name:       "Room",
		Visibility: "public",
		HostUserID: 1,
		InviteLink: "invite-123",
		Members: []domain.RoomMember{
			{UserID: 1, Role: "host", Status: "active"},
			{UserID: 2, Role: "member", Status: "active"},
		},
	}

	return &roomUsecaseRepo{
		room: room,
		overview: domain.OverviewResponse{
			ActiveRooms: []domain.RoomCard{
				{ID: 5, HostUserID: 1, InviteLink: "invite-123"},
			},
			MyRooms: []domain.RoomCard{
				{ID: 6, HostUserID: 7, InviteLink: "invite-456"},
			},
			FeaturedRooms: []domain.RoomCard{
				{ID: 8, HostUserID: 3, InviteLink: "invite-789"},
			},
		},
		invite: &domain.Invite{
			RoomID: 5,
			Link:   "invite-123",
		},
	}
}

func (r *roomUsecaseRepo) GetOverview(context.Context, int64) (domain.OverviewResponse, error) {
	return r.overview, nil
}

func (r *roomUsecaseRepo) GetRoomByID(_ context.Context, roomID int64) (*domain.Room, error) {
	if r.room == nil || r.room.ID != roomID {
		return nil, domain.ErrRoomNotFound
	}

	roomCopy := *r.room
	roomCopy.Members = append([]domain.RoomMember(nil), r.room.Members...)

	return &roomCopy, nil
}

func (r *roomUsecaseRepo) CreateRoom(context.Context, int64, domain.CreateRoomRequest) (*domain.Room, error) {
	return nil, domain.ErrNotImplemented
}

func (r *roomUsecaseRepo) InviteMember(_ context.Context, roomID, userID int64) error {
	r.inviteMemberRoomID = roomID
	r.inviteMemberUserID = userID

	for i := range r.room.Members {
		if r.room.Members[i].UserID == userID {
			return nil
		}
	}

	r.room.Members = append(r.room.Members, domain.RoomMember{UserID: userID, Role: "member", Status: "pending"})

	return nil
}

func (r *roomUsecaseRepo) AddMember(_ context.Context, roomID, userID int64) (*domain.Room, error) {
	r.addMemberRoomID = roomID
	r.addMemberUserID = userID

	roomCopy := *r.room
	roomCopy.Members = append([]domain.RoomMember(nil), r.room.Members...)
	roomCopy.Members = append(roomCopy.Members, domain.RoomMember{UserID: userID, Role: "member", Status: "active"})

	return &roomCopy, nil
}

func (r *roomUsecaseRepo) ActivateMember(_ context.Context, roomID, userID int64) error {
	r.activateMemberRoomID = roomID
	r.activateMemberUserID = userID

	for i := range r.room.Members {
		if r.room.Members[i].UserID == userID {
			r.room.Members[i].Status = "active"
		}
	}

	return nil
}

func (r *roomUsecaseRepo) DeleteRoom(context.Context, int64) error {
	return domain.ErrNotImplemented
}

func (r *roomUsecaseRepo) GetInvite(_ context.Context, inviteLink string) (*domain.Invite, error) {
	if r.invite == nil || r.invite.Link != inviteLink {
		return nil, domain.ErrInviteNotFound
	}

	return r.invite, nil
}

func (r *roomUsecaseRepo) SavePlaybackState(context.Context, int64, domain.PlaybackState) error {
	return domain.ErrNotImplemented
}

func (r *roomUsecaseRepo) SaveMessage(context.Context, domain.RoomMessage) (*domain.RoomMessage, error) {
	return nil, domain.ErrNotImplemented
}

func (r *roomUsecaseRepo) SavePoll(context.Context, domain.Poll) (*domain.Poll, error) {
	return nil, domain.ErrNotImplemented
}

func (r *roomUsecaseRepo) SaveVote(context.Context, domain.PollVote) error {
	return domain.ErrNotImplemented
}

func (r *roomUsecaseRepo) TouchRoom(context.Context, int64) error {
	return domain.ErrNotImplemented
}

func findMemberStatus(t *testing.T, members []domain.RoomMember, userID int64) string {
	t.Helper()

	for _, member := range members {
		if member.UserID == userID {
			return member.Status
		}
	}

	t.Fatalf("member %d not found", userID)

	return ""
}

type roomUsecaseBroker struct {
	subscribedRoomIDs []int64
}

func (b *roomUsecaseBroker) Publish(context.Context, domain.RoomEvent) error {
	return nil
}

func (b *roomUsecaseBroker) Subscribe(_ context.Context, roomID int64) (<-chan domain.RoomEvent, func(), error) {
	b.subscribedRoomIDs = append(b.subscribedRoomIDs, roomID)

	ch := make(chan domain.RoomEvent)

	return ch, func() {
		close(ch)
	}, nil
}
