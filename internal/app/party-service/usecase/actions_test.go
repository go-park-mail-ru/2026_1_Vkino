package usecase

import (
	"context"
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
		{action: roomActionPlay, status: "playing"},
		{action: roomActionPause, status: playbackStatusPaused},
	} {
		t.Run(tc.action, func(t *testing.T) {
			t.Parallel()

			repo := newPlaybackActionRepo()
			broker := &playbackActionBroker{}
			svc := New(repo, broker, nil, nil)

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
	svc := New(repo, broker, nil, nil)

	playback, err := svc.ApplyRoomAction(context.Background(), 2, domain.ApplyRoomActionRequest{
		RoomID:          5,
		Action:          roomActionSeek,
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
	require.Equal(t, roomActionSeek, broker.events[0].Type)
	require.Equal(t, int64(2), broker.events[0].ActorUserID)
	require.NotNil(t, broker.events[0].Playback)
	require.Equal(t, int64(40), broker.events[0].Playback.PositionSeconds)
}

func TestApplyRoomActionAllowsNonHostParticipantSyncState(t *testing.T) {
	t.Parallel()

	repo := newPlaybackActionRepo()
	broker := &playbackActionBroker{}
	svc := New(repo, broker, nil, nil)

	playback, err := svc.ApplyRoomAction(context.Background(), 2, domain.ApplyRoomActionRequest{
		RoomID:          5,
		Action:          roomActionSyncState,
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
	require.Equal(t, roomActionSyncState, broker.events[0].Type)
	require.Equal(t, int64(2), broker.events[0].ActorUserID)
	require.NotNil(t, broker.events[0].Playback)
	require.Equal(t, "paused", broker.events[0].Playback.Status)
}

func TestApplyRoomActionRejectsNonMember(t *testing.T) {
	t.Parallel()

	repo := newPlaybackActionRepo()
	broker := &playbackActionBroker{}
	svc := New(repo, broker, nil, nil)

	_, err := svc.ApplyRoomAction(context.Background(), 3, domain.ApplyRoomActionRequest{
		RoomID:          5,
		Action:          roomActionPlay,
		PositionSeconds: 26,
		DurationSeconds: 70,
	})

	require.ErrorIs(t, err, domain.ErrAccessDenied)
	require.Empty(t, repo.savedPlayback)
	require.Empty(t, repo.touchedRooms)
	require.Empty(t, broker.events)
}

func TestApplyRoomActionKeepsMovieSelectionHostOnly(t *testing.T) {
	t.Parallel()

	repo := newPlaybackActionRepo()
	broker := &playbackActionBroker{}
	svc := New(repo, broker, nil, nil)

	_, err := svc.ApplyRoomAction(context.Background(), 2, domain.ApplyRoomActionRequest{
		RoomID:  5,
		Action:  "select_movie",
		MovieID: 9,
	})

	require.ErrorIs(t, err, domain.ErrAccessDenied)
	require.Empty(t, repo.savedPlayback)
	require.Empty(t, repo.touchedRooms)
	require.Empty(t, broker.events)
}

func TestVoteRoomPollSpendsCoinsAndUpdatesPoll(t *testing.T) {
	t.Parallel()

	repo := newVotePollRepo()
	broker := &playbackActionBroker{}
	spender := &votePollCoinsSpender{}
	svc := New(repo, broker, nil, spender)

	vote, poll, err := svc.VoteRoomPoll(context.Background(), 2, domain.VoteRoomPollRequest{
		RoomID:      5,
		PollID:      100,
		OptionID:    1001,
		CoinsAmount: 25,
	})

	require.NoError(t, err)
	require.Equal(t, int32(25), vote.CoinsAmount)
	require.Len(t, repo.savedVotes, 1)
	require.Equal(t, vote, repo.savedVotes[0])
	require.Equal(t, int32(25), spender.coinsAmount)
	require.Equal(t, int64(5), spender.roomID)
	require.Equal(t, int64(100), spender.pollID)
	require.Equal(t, int64(1001), spender.optionID)
	require.Equal(t, int64(2), spender.userID)
	require.Equal(t, int64(1), poll.Options[0].VotesCount)
	require.Equal(t, int64(25), poll.Options[0].CoinsTotal)
	require.Len(t, broker.events, 1)
	require.NotNil(t, broker.events[0].Vote)
	require.Equal(t, int32(25), broker.events[0].Vote.CoinsAmount)
}

func TestVoteRoomPollStopsWhenCoinsSpendFails(t *testing.T) {
	t.Parallel()

	repo := newVotePollRepo()
	broker := &playbackActionBroker{}
	spender := &votePollCoinsSpender{err: domain.ErrInsufficientVKinoCoins}
	svc := New(repo, broker, nil, spender)

	_, _, err := svc.VoteRoomPoll(context.Background(), 2, domain.VoteRoomPollRequest{
		RoomID:      5,
		PollID:      100,
		OptionID:    1001,
		CoinsAmount: 25,
	})

	require.ErrorIs(t, err, domain.ErrInsufficientVKinoCoins)
	require.Empty(t, repo.savedVotes)
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
				{UserID: 1, Role: memberRoleHost, Status: memberStatusActive},
				{UserID: 2, Role: memberRoleMember, Status: memberStatusActive},
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

type votePollRepo struct {
	room       *domain.Room
	savedVotes []domain.PollVote
}

func newVotePollRepo() *votePollRepo {
	return &votePollRepo{
		room: &domain.Room{
			ID:         5,
			Name:       "Room",
			Visibility: "public",
			HostUserID: 1,
			Members: []domain.RoomMember{
				{UserID: 1, Role: memberRoleHost, Status: memberStatusActive},
				{UserID: 2, Role: memberRoleMember, Status: memberStatusActive},
			},
			Polls: []domain.Poll{
				{
					ID:              100,
					RoomID:          5,
					Question:        "Who wins?",
					CreatedByUserID: 1,
					Options: []domain.PollOption{
						{ID: 1001, Title: "A"},
						{ID: 1002, Title: "B"},
					},
				},
			},
		},
	}
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

func (r *votePollRepo) GetOverview(context.Context, int64) (domain.OverviewResponse, error) {
	return domain.OverviewResponse{}, nil
}

func (r *votePollRepo) GetRoomByID(_ context.Context, roomID int64) (*domain.Room, error) {
	if r.room == nil || r.room.ID != roomID {
		return nil, domain.ErrRoomNotFound
	}

	return r.room, nil
}

func (r *votePollRepo) CreateRoom(context.Context, int64, domain.CreateRoomRequest) (*domain.Room, error) {
	return nil, domain.ErrNotImplemented
}

func (r *votePollRepo) InviteMember(context.Context, int64, int64) error {
	return domain.ErrNotImplemented
}

func (r *votePollRepo) AddMember(context.Context, int64, int64) (*domain.Room, error) {
	return nil, domain.ErrNotImplemented
}

func (r *votePollRepo) ActivateMember(context.Context, int64, int64) error {
	return nil
}

func (r *votePollRepo) DeleteRoom(context.Context, int64) error {
	return domain.ErrNotImplemented
}

func (r *votePollRepo) GetInvite(context.Context, string) (*domain.Invite, error) {
	return nil, domain.ErrNotImplemented
}

func (r *votePollRepo) SavePlaybackState(context.Context, int64, domain.PlaybackState) error {
	return domain.ErrNotImplemented
}

func (r *votePollRepo) SaveMessage(context.Context, domain.RoomMessage) (*domain.RoomMessage, error) {
	return nil, domain.ErrNotImplemented
}

func (r *votePollRepo) SavePoll(context.Context, domain.Poll) (*domain.Poll, error) {
	return nil, domain.ErrNotImplemented
}

func (r *votePollRepo) SaveVote(_ context.Context, vote domain.PollVote) error {
	r.savedVotes = append(r.savedVotes, vote)
	for pollIdx := range r.room.Polls {
		for optionIdx := range r.room.Polls[pollIdx].Options {
			if r.room.Polls[pollIdx].Options[optionIdx].ID == vote.OptionID {
				r.room.Polls[pollIdx].Options[optionIdx].VotesCount++
				r.room.Polls[pollIdx].Options[optionIdx].CoinsTotal += int64(vote.CoinsAmount)
			}
		}
	}

	return nil
}

func (r *votePollRepo) TouchRoom(context.Context, int64) error {
	return nil
}

func (r *playbackActionRepo) TouchRoom(_ context.Context, roomID int64) error {
	r.touchedRooms = append(r.touchedRooms, roomID)

	return nil
}

type playbackActionBroker struct {
	events []domain.RoomEvent
}

type votePollCoinsSpender struct {
	userID      int64
	roomID      int64
	pollID      int64
	optionID    int64
	coinsAmount int32
	err         error
}

func (s *votePollCoinsSpender) SpendForPollVote(
	_ context.Context,
	userID, roomID, pollID, optionID int64,
	coinsAmount int32,
) error {
	s.userID = userID
	s.roomID = roomID
	s.pollID = pollID
	s.optionID = optionID
	s.coinsAmount = coinsAmount

	return s.err
}

func (b *playbackActionBroker) Publish(_ context.Context, event domain.RoomEvent) error {
	b.events = append(b.events, event)

	return nil
}

func (b *playbackActionBroker) Subscribe(
	context.Context,
	int64,
	int64,
) (<-chan domain.RoomEvent, func(), error) {
	return nil, nil, domain.ErrNotImplemented
}

func (b *playbackActionBroker) ActiveUsers(int64) int32 {
	return 0
}

func (b *playbackActionBroker) IsUserActive(int64, int64) bool {
	return false
}
