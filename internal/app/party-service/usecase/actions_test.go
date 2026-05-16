package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	"github.com/stretchr/testify/require"
)

func TestApplyRoomActionAllowsNonHostParticipantPlaybackActions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		action string
		status string
	}{
		{action: "play", status: "playing"},
		{action: "pause", status: "paused"},
	} {
		t.Run(tc.action, func(t *testing.T) {
			t.Parallel()

			repo := newPlaybackActionRepo()
			broker := &playbackActionBroker{}
			svc := New(repo, broker)

			playback, err := svc.ApplyRoomAction(context.Background(), 2, domain.ApplyRoomActionRequest{
				RoomID:          5,
				Action:          tc.action,
				EpisodeID:       8,
				DurationSeconds: 70,
				PositionSeconds: 26,
			})

			require.NoError(t, err)
			require.Equal(t, int64(8), playback.EpisodeID)
			require.Equal(t, int64(26), playback.PositionSeconds)
			require.Equal(t, int64(70), playback.DurationSeconds)
			require.Equal(t, tc.status, playback.Status)

			require.Len(t, repo.savedPlayback, 1)
			require.Equal(t, playback, repo.savedPlayback[0])
			require.Equal(t, []int64{5}, repo.touchedRooms)

			require.Len(t, broker.events, 1)
			require.Equal(t, tc.action, broker.events[0].Type)
			require.Equal(t, int64(5), broker.events[0].RoomID)
			require.Equal(t, int64(2), broker.events[0].ActorUserID)
			require.NotNil(t, broker.events[0].Playback)
			require.Equal(t, tc.status, broker.events[0].Playback.Status)
		})
	}
}

func TestApplyRoomActionAllowsNonHostParticipantSeek(t *testing.T) {
	t.Parallel()

	repo := newPlaybackActionRepo()
	broker := &playbackActionBroker{}
	svc := New(repo, broker)

	playback, err := svc.ApplyRoomAction(context.Background(), 2, domain.ApplyRoomActionRequest{
		RoomID:          5,
		Action:          "seek",
		DurationSeconds: 70,
		PositionSeconds: 40,
	})

	require.NoError(t, err)
	require.Equal(t, int64(40), playback.PositionSeconds)
	require.Equal(t, int64(70), playback.DurationSeconds)
	require.Equal(t, "paused", playback.Status)

	require.Len(t, repo.savedPlayback, 1)
	require.Equal(t, playback, repo.savedPlayback[0])

	require.Len(t, broker.events, 1)
	require.Equal(t, "seek", broker.events[0].Type)
	require.Equal(t, int64(2), broker.events[0].ActorUserID)
	require.NotNil(t, broker.events[0].Playback)
	require.Equal(t, int64(40), broker.events[0].Playback.PositionSeconds)
}

func TestApplyRoomActionAllowsNonHostParticipantSyncState(t *testing.T) {
	t.Parallel()

	repo := newPlaybackActionRepo()
	broker := &playbackActionBroker{}
	svc := New(repo, broker)

	playback, err := svc.ApplyRoomAction(context.Background(), 2, domain.ApplyRoomActionRequest{
		RoomID:          5,
		Action:          "sync_state",
		MovieID:         11,
		EpisodeID:       12,
		DurationSeconds: 70,
		PositionSeconds: 0,
		Status:          "paused",
	})

	require.NoError(t, err)
	require.Equal(t, int64(11), playback.MovieID)
	require.Equal(t, int64(12), playback.EpisodeID)
	require.Equal(t, int64(0), playback.PositionSeconds)
	require.Equal(t, int64(70), playback.DurationSeconds)
	require.Equal(t, "paused", playback.Status)

	require.Len(t, repo.savedPlayback, 1)
	require.Equal(t, playback, repo.savedPlayback[0])

	require.Len(t, broker.events, 1)
	require.Equal(t, "sync_state", broker.events[0].Type)
	require.Equal(t, int64(2), broker.events[0].ActorUserID)
	require.NotNil(t, broker.events[0].Playback)
	require.Equal(t, "paused", broker.events[0].Playback.Status)
}

func TestApplyRoomActionRejectsNonMember(t *testing.T) {
	t.Parallel()

	repo := newPlaybackActionRepo()
	broker := &playbackActionBroker{}
	svc := New(repo, broker)

	_, err := svc.ApplyRoomAction(context.Background(), 3, domain.ApplyRoomActionRequest{
		RoomID:          5,
		Action:          "play",
		PositionSeconds: 26,
		DurationSeconds: 70,
	})

	require.True(t, errors.Is(err, domain.ErrAccessDenied))
	require.Empty(t, repo.savedPlayback)
	require.Empty(t, repo.touchedRooms)
	require.Empty(t, broker.events)
}

func TestApplyRoomActionKeepsMovieSelectionHostOnly(t *testing.T) {
	t.Parallel()

	repo := newPlaybackActionRepo()
	broker := &playbackActionBroker{}
	svc := New(repo, broker)

	_, err := svc.ApplyRoomAction(context.Background(), 2, domain.ApplyRoomActionRequest{
		RoomID:  5,
		Action:  "select_movie",
		MovieID: 9,
	})

	require.True(t, errors.Is(err, domain.ErrAccessDenied))
	require.Empty(t, repo.savedPlayback)
	require.Empty(t, repo.touchedRooms)
	require.Empty(t, broker.events)
}

func newPlaybackActionRepo() *playbackActionRepo {
	return &playbackActionRepo{
		room: &domain.Room{
			ID:         5,
			Name:       "Room",
			Visibility: "public",
			HostUserID: 1,
			Members: []domain.RoomMember{
				{UserID: 1, Role: "host", Status: "active"},
				{UserID: 2, Role: "member", Status: "active"},
			},
			Playback: domain.PlaybackState{
				MovieID:         7,
				EpisodeID:       6,
				DurationSeconds: 60,
				PositionSeconds: 10,
				Status:          "paused",
			},
		},
	}
}

type playbackActionRepo struct {
	room          *domain.Room
	savedPlayback []domain.PlaybackState
	touchedRooms  []int64
}

func (r *playbackActionRepo) GetOverview(context.Context, int64) (domain.OverviewResponse, error) {
	return domain.OverviewResponse{}, nil
}

func (r *playbackActionRepo) GetRoomByID(_ context.Context, roomID int64) (*domain.Room, error) {
	if r.room == nil || r.room.ID != roomID {
		return nil, domain.ErrRoomNotFound
	}

	return r.room, nil
}

func (r *playbackActionRepo) CreateRoom(
	context.Context,
	int64,
	domain.CreateRoomRequest,
) (*domain.Room, error) {
	return nil, domain.ErrNotImplemented
}

func (r *playbackActionRepo) InviteMember(context.Context, int64, int64) error {
	return domain.ErrNotImplemented
}

func (r *playbackActionRepo) AddMember(context.Context, int64, int64) (*domain.Room, error) {
	return nil, domain.ErrNotImplemented
}

func (r *playbackActionRepo) ActivateMember(context.Context, int64, int64) error {
	return nil
}

func (r *playbackActionRepo) DeleteRoom(context.Context, int64) error {
	return domain.ErrNotImplemented
}

func (r *playbackActionRepo) GetInvite(context.Context, string) (*domain.Invite, error) {
	return nil, domain.ErrNotImplemented
}

func (r *playbackActionRepo) SavePlaybackState(_ context.Context, _ int64, state domain.PlaybackState) error {
	r.savedPlayback = append(r.savedPlayback, state)
	r.room.Playback = state

	return nil
}

func (r *playbackActionRepo) SaveMessage(
	context.Context,
	domain.RoomMessage,
) (*domain.RoomMessage, error) {
	return nil, domain.ErrNotImplemented
}

func (r *playbackActionRepo) SavePoll(context.Context, domain.Poll) (*domain.Poll, error) {
	return nil, domain.ErrNotImplemented
}

func (r *playbackActionRepo) SaveVote(context.Context, domain.PollVote) error {
	return domain.ErrNotImplemented
}

func (r *playbackActionRepo) TouchRoom(_ context.Context, roomID int64) error {
	r.touchedRooms = append(r.touchedRooms, roomID)

	return nil
}

type playbackActionBroker struct {
	events []domain.RoomEvent
}

func (b *playbackActionBroker) Publish(_ context.Context, event domain.RoomEvent) error {
	b.events = append(b.events, event)

	return nil
}

func (b *playbackActionBroker) Subscribe(
	context.Context,
	int64,
) (<-chan domain.RoomEvent, func(), error) {
	return nil, nil, domain.ErrNotImplemented
}
