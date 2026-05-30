package usecase

import (
	"context"
	"slices"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	"github.com/stretchr/testify/require"
)

const testInviteLink = "invite-123"

func TestGetRoomAutoJoinsPublicRoomForNonMember(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

	resp, err := svc.GetRoom(context.Background(), 3, 5)

	require.NoError(t, err)
	require.Equal(t, int64(5), repo.addMemberRoomID)
	require.Equal(t, int64(3), repo.addMemberUserID)
	require.Equal(t, memberStatusActive, findMemberStatus(t, resp.Room.Members, 3))
}

func TestGetRoomHidesInviteForNonHost(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

	resp, err := svc.GetRoom(context.Background(), 2, 5)
	require.NoError(t, err)
	require.Empty(t, resp.Room.InviteLink)
}

func TestGetRoomInviteAllowsHostOnly(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

	resp, err := svc.GetRoomInvite(context.Background(), 1, 5)
	require.NoError(t, err)
	require.Equal(t, int64(5), resp.RoomID)
	require.Equal(t, testInviteLink, resp.InviteLink)
}

func TestInviteFriendToRoomCreatesPendingMember(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

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
	repo.room.Members = append(repo.room.Members, domain.RoomMember{
		UserID: 9, Role: memberRoleMember, Status: memberStatusPending,
	})
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

	resp, err := svc.GetRoom(context.Background(), 9, 5)
	require.NoError(t, err)
	require.Equal(t, int64(5), repo.activateMemberRoomID)
	require.Equal(t, int64(9), repo.activateMemberUserID)
	require.Equal(t, memberStatusActive, findMemberStatus(t, resp.Room.Members, 9))
}

func TestJoinRoomUsesInviteOnly(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

	_, err := svc.JoinRoom(context.Background(), 9, domain.JoinRoomRequest{})

	require.ErrorIs(t, err, domain.ErrInvalidInviteLink)
}

func TestJoinRoomAddsMemberByInviteAndHidesInviteForNonHost(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

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
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

	resp, err := svc.GetOverview(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, testInviteLink, resp.ActiveRooms[0].InviteLink)
	require.Empty(t, resp.FeaturedRooms[0].InviteLink)
	require.Empty(t, resp.MyRooms[0].InviteLink)
}

func TestSubscribeRoomRequiresMembership(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	broker := &roomUsecaseBroker{}
	svc := New(repo, broker, nil, nil)

	_, _, err := svc.SubscribeRoom(context.Background(), 3, domain.SubscribeRoomRequest{RoomID: 5})

	require.ErrorIs(t, err, domain.ErrAccessDenied)
	require.Empty(t, broker.subscribedRoomIDs)
}

func TestCreateRoomNormalizesVisibility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		visibility string
		want       string
	}{
		{name: "public russian", visibility: "Открытая", want: "public"},
		{name: "public legacy", visibility: "open", want: "public"},
		{name: "private russian", visibility: "Закрытая", want: "private"},
		{name: "private canonical", visibility: "private", want: "private"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newRoomUsecaseRepo()
			svc := New(repo, &roomUsecaseBroker{}, nil, nil)

			resp, err := svc.CreateRoom(context.Background(), 1, domain.CreateRoomRequest{
				Name:       "  Room  ",
				Visibility: tt.visibility,
			})
			require.NoError(t, err)
			require.Equal(t, tt.want, resp.Room.Visibility)
			require.Equal(t, "Room", repo.createdRoomReq.Name)
			require.Equal(t, tt.want, repo.createdRoomReq.Visibility)
		})
	}
}

func TestCreateRoomRejectsUnknownVisibility(t *testing.T) {
	t.Parallel()

	repo := newRoomUsecaseRepo()
	svc := New(repo, &roomUsecaseBroker{}, nil, nil)

	_, err := svc.CreateRoom(context.Background(), 1, domain.CreateRoomRequest{
		Name:       "Room",
		Visibility: "friends-only",
	})

	require.ErrorIs(t, err, domain.ErrInvalidVisibility)
	require.False(t, repo.createRoomCalled)
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
	createdRoomReq       domain.CreateRoomRequest
	createRoomCalled     bool
}

func newRoomUsecaseRepo() *roomUsecaseRepo {
	room := &domain.Room{
		ID:         5,
		Name:       "Room",
		Visibility: "public",
		HostUserID: 1,
		InviteLink: testInviteLink,
		Members: []domain.RoomMember{
			{UserID: 1, Role: memberRoleHost, Status: memberStatusActive},
			{UserID: 2, Role: memberRoleMember, Status: memberStatusActive},
		},
	}

	return &roomUsecaseRepo{
		room: room,
		overview: domain.OverviewResponse{
			ActiveRooms: []domain.RoomCard{
				{ID: 5, HostUserID: 1, InviteLink: testInviteLink},
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
			Link:   testInviteLink,
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

func (r *roomUsecaseRepo) CreateRoom(
	_ context.Context,
	userID int64,
	req domain.CreateRoomRequest,
) (*domain.Room, error) {
	r.createRoomCalled = true
	r.createdRoomReq = req

	return &domain.Room{
		ID:         99,
		Name:       req.Name,
		Visibility: req.Visibility,
		HostUserID: userID,
		InviteLink: "invite-created",
		Members: []domain.RoomMember{
			{UserID: userID, Role: memberRoleHost, Status: memberStatusActive},
		},
	}, nil
}

func (r *roomUsecaseRepo) InviteMember(_ context.Context, roomID, userID int64) error {
	r.inviteMemberRoomID = roomID
	r.inviteMemberUserID = userID

	for i := range r.room.Members {
		if r.room.Members[i].UserID == userID {
			return nil
		}
	}

	r.room.Members = append(r.room.Members, domain.RoomMember{
		UserID: userID, Role: memberRoleMember, Status: memberStatusPending,
	})

	return nil
}

func (r *roomUsecaseRepo) AddMember(_ context.Context, roomID, userID int64) (*domain.Room, error) {
	r.addMemberRoomID = roomID
	r.addMemberUserID = userID

	roomCopy := *r.room
	roomCopy.Members = append([]domain.RoomMember(nil), r.room.Members...)
	roomCopy.Members = append(roomCopy.Members, domain.RoomMember{
		UserID: userID, Role: memberRoleMember, Status: memberStatusActive,
	})

	return &roomCopy, nil
}

func (r *roomUsecaseRepo) ActivateMember(_ context.Context, roomID, userID int64) error {
	r.activateMemberRoomID = roomID
	r.activateMemberUserID = userID

	for i := range r.room.Members {
		if r.room.Members[i].UserID == userID {
			r.room.Members[i].Status = memberStatusActive
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

func (r *roomUsecaseRepo) GetPollOptionStakes(context.Context, int64, int64) ([]domain.PollOptionStake, error) {
	return nil, domain.ErrNotImplemented
}

func (r *roomUsecaseRepo) ResolvePoll(context.Context, int64, int64, int64, int64) (*domain.Poll, error) {
	return nil, domain.ErrNotImplemented
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

func (b *roomUsecaseBroker) Subscribe(_ context.Context, roomID, _ int64) (<-chan domain.RoomEvent, func(), error) {
	b.subscribedRoomIDs = append(b.subscribedRoomIDs, roomID)

	ch := make(chan domain.RoomEvent)

	return ch, func() {
		close(ch)
	}, nil
}

func (b *roomUsecaseBroker) ActiveUsers(roomID int64) int32 {
	if slices.Contains(b.subscribedRoomIDs, roomID) {
		return 1
	}

	return 0
}

func (b *roomUsecaseBroker) IsUserActive(int64, int64) bool {
	return false
}
