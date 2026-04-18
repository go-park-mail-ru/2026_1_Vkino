package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase/mocks"
	"go.uber.org/mock/gomock"
)

func newMovieUsecase(
	repo *mocks.MockMovieRepo,
	actorStorage *mocks.MockFileStorage,
	posterStorage *mocks.MockFileStorage,
	cardStorage *mocks.MockFileStorage,
	videoStorage *mocks.MockFileStorage,
) *usecase.MovieUsecase {
	return usecase.NewMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)
}

func newMovieStorages(
	ctrl *gomock.Controller,
) (*mocks.MockFileStorage, *mocks.MockFileStorage, *mocks.MockFileStorage, *mocks.MockFileStorage) {
	return mocks.NewMockFileStorage(ctrl), mocks.NewMockFileStorage(ctrl), mocks.NewMockFileStorage(ctrl), mocks.NewMockFileStorage(ctrl)
}

func TestNewMovieUsecase(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockMovieRepo(ctrl)
	actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)

	if got := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage); got == nil {
		t.Fatal("expected non-nil usecase")
	}
}

func TestMovieUsecase_GetAllSelections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func(repo *mocks.MockMovieRepo)
		want      []domain.SelectionResponse
		wantErr   error
	}{
		{
			name: "repo error",
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					GetAllSelections(gomock.Any()).
					Return(nil, errors.New("repo failed"))
			},
			wantErr: errors.New("repo failed"),
		},
		{
			name: "success",
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					GetAllSelections(gomock.Any()).
					Return([]domain.SelectionResponse{{Title: "popular"}}, nil)
			},
			want: []domain.SelectionResponse{{Title: "popular"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMovieRepo(ctrl)
			actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)
			tt.setupMock(repo)

			u := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)
			got, err := u.GetAllSelections(context.Background())

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %q, got %v", tt.wantErr.Error(), err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) || got[0].Title != tt.want[0].Title {
				t.Fatalf("expected %#v, got %#v", tt.want, got)
			}
		})
	}
}

func TestMovieUsecase_GetSelectionByTitle(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockMovieRepo(ctrl)
	actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)
	want := domain.SelectionResponse{Title: "popular"}

	repo.EXPECT().
		GetSelectionByTitle(gomock.Any(), "popular").
		Return(want, nil)

	u := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)

	got, err := u.GetSelectionByTitle(context.Background(), "popular")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Title != want.Title {
		t.Fatalf("expected title %q, got %q", want.Title, got.Title)
	}
}

func TestMovieUsecase_GetMovieByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        int64
		setupMock func(repo *mocks.MockMovieRepo)
		wantErr   error
		wantMovie domain.MovieResponse
	}{
		{
			name:    "invalid id",
			id:      0,
			wantErr: domain.ErrInvalidMovieID,
		},
		{
			name: "movie repo error",
			id:   7,
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					GetMovieByID(gomock.Any(), int64(7)).
					Return(domain.MovieResponse{}, errors.New("movie failed"))
			},
			wantErr: errors.New("movie failed"),
		},
		{
			name: "episodes repo error",
			id:   7,
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					GetMovieByID(gomock.Any(), int64(7)).
					Return(domain.MovieResponse{ID: 7, Title: "Dune"}, nil)
				repo.EXPECT().
					GetEpisodesByMovieID(gomock.Any(), int64(7)).
					Return(nil, errors.New("episodes failed"))
			},
			wantErr: errors.New("episodes failed"),
		},
		{
			name: "success",
			id:   7,
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					GetMovieByID(gomock.Any(), int64(7)).
					Return(domain.MovieResponse{ID: 7, Title: "Dune"}, nil)
				repo.EXPECT().
					GetEpisodesByMovieID(gomock.Any(), int64(7)).
					Return([]domain.EpisodeItemResponse{{ID: 1, Title: "Episode 1"}}, nil)
			},
			wantMovie: domain.MovieResponse{
				ID:       7,
				Title:    "Dune",
				Episodes: []domain.EpisodeItemResponse{{ID: 1, Title: "Episode 1"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMovieRepo(ctrl)
			actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			u := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)
			got, err := u.GetMovieByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) && (err == nil || err.Error() != tt.wantErr.Error()) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.ID != tt.wantMovie.ID || got.Title != tt.wantMovie.Title || len(got.Episodes) != 1 {
				t.Fatalf("expected movie %#v, got %#v", tt.wantMovie, got)
			}
		})
	}
}

func TestMovieUsecase_GetActorByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockMovieRepo(ctrl)
	actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)
	want := domain.ActorResponse{ID: 9, FullName: "Actor"}

	repo.EXPECT().
		GetActorByID(gomock.Any(), int64(9)).
		Return(want, nil)

	u := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)

	got, err := u.GetActorByID(context.Background(), 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != want.ID || got.FullName != want.FullName {
		t.Fatalf("expected actor %#v, got %#v", want, got)
	}
}

func TestMovieUsecase_GetEpisodePlayback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		episodeID int64
		userID    int64
		setupMock func(repo *mocks.MockMovieRepo, videoStorage *mocks.MockFileStorage)
		wantErr   error
		want      domain.EpisodePlaybackResponse
	}{
		{
			name:      "invalid episode id",
			episodeID: 0,
			wantErr:   domain.ErrInvalidEpisodeID,
		},
		{
			name:      "repo error",
			episodeID: 4,
			setupMock: func(repo *mocks.MockMovieRepo, videoStorage *mocks.MockFileStorage) {
				repo.EXPECT().
					GetEpisodePlayback(gomock.Any(), int64(4)).
					Return(domain.EpisodePlaybackResponse{}, errors.New("playback failed"))
			},
			wantErr: errors.New("playback failed"),
		},
		{
			name:      "presign error",
			episodeID: 4,
			setupMock: func(repo *mocks.MockMovieRepo, videoStorage *mocks.MockFileStorage) {
				repo.EXPECT().
					GetEpisodePlayback(gomock.Any(), int64(4)).
					Return(domain.EpisodePlaybackResponse{EpisodeID: 4, PlaybackURL: "video/key.mp4"}, nil)
				videoStorage.EXPECT().
					PresignGetObject(gomock.Any(), "video/key.mp4", time.Duration(0)).
					Return("", errors.New("presign failed"))
			},
			wantErr: errors.New("presign failed"),
		},
		{
			name:      "watch progress error",
			episodeID: 4,
			userID:    5,
			setupMock: func(repo *mocks.MockMovieRepo, videoStorage *mocks.MockFileStorage) {
				repo.EXPECT().
					GetEpisodePlayback(gomock.Any(), int64(4)).
					Return(domain.EpisodePlaybackResponse{EpisodeID: 4, PlaybackURL: "video/key.mp4"}, nil)
				videoStorage.EXPECT().
					PresignGetObject(gomock.Any(), "video/key.mp4", time.Duration(0)).
					Return("https://cdn.example/video.mp4", nil)
				repo.EXPECT().
					GetWatchProgress(gomock.Any(), int64(5), int64(4)).
					Return(0, errors.New("progress failed"))
			},
			wantErr: errors.New("progress failed"),
		},
		{
			name:      "success anonymous",
			episodeID: 4,
			setupMock: func(repo *mocks.MockMovieRepo, videoStorage *mocks.MockFileStorage) {
				repo.EXPECT().
					GetEpisodePlayback(gomock.Any(), int64(4)).
					Return(domain.EpisodePlaybackResponse{EpisodeID: 4, PlaybackURL: "video/key.mp4"}, nil)
				videoStorage.EXPECT().
					PresignGetObject(gomock.Any(), "video/key.mp4", time.Duration(0)).
					Return("https://cdn.example/video.mp4", nil)
			},
			want: domain.EpisodePlaybackResponse{EpisodeID: 4, PlaybackURL: "https://cdn.example/video.mp4"},
		},
		{
			name:      "success with progress",
			episodeID: 4,
			userID:    5,
			setupMock: func(repo *mocks.MockMovieRepo, videoStorage *mocks.MockFileStorage) {
				repo.EXPECT().
					GetEpisodePlayback(gomock.Any(), int64(4)).
					Return(domain.EpisodePlaybackResponse{EpisodeID: 4, PlaybackURL: "video/key.mp4"}, nil)
				videoStorage.EXPECT().
					PresignGetObject(gomock.Any(), "video/key.mp4", time.Duration(0)).
					Return("https://cdn.example/video.mp4", nil)
				repo.EXPECT().
					GetWatchProgress(gomock.Any(), int64(5), int64(4)).
					Return(77, nil)
			},
			want: domain.EpisodePlaybackResponse{
				EpisodeID:       4,
				PlaybackURL:     "https://cdn.example/video.mp4",
				PositionSeconds: 77,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMovieRepo(ctrl)
			actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(repo, videoStorage)
			}

			u := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)
			got, err := u.GetEpisodePlayback(context.Background(), tt.episodeID, tt.userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) && (err == nil || err.Error() != tt.wantErr.Error()) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.EpisodeID != tt.want.EpisodeID || got.PlaybackURL != tt.want.PlaybackURL ||
				got.PositionSeconds != tt.want.PositionSeconds {
				t.Fatalf("expected %#v, got %#v", tt.want, got)
			}
		})
	}
}

func TestMovieUsecase_GetEpisodeProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		episodeID int64
		setupMock func(repo *mocks.MockMovieRepo)
		wantErr   error
		want      domain.WatchProgressResponse
	}{
		{
			name:      "invalid episode id",
			episodeID: 0,
			wantErr:   domain.ErrInvalidEpisodeID,
		},
		{
			name:      "repo error",
			episodeID: 8,
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					GetWatchProgress(gomock.Any(), int64(11), int64(8)).
					Return(0, errors.New("progress failed"))
			},
			wantErr: errors.New("progress failed"),
		},
		{
			name:      "success",
			episodeID: 8,
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					GetWatchProgress(gomock.Any(), int64(11), int64(8)).
					Return(33, nil)
			},
			want: domain.WatchProgressResponse{EpisodeID: 8, PositionSeconds: 33},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMovieRepo(ctrl)
			actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			u := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)
			got, err := u.GetEpisodeProgress(context.Background(), 11, tt.episodeID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) && (err == nil || err.Error() != tt.wantErr.Error()) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("expected %#v, got %#v", tt.want, got)
			}
		})
	}
}

func TestMovieUsecase_SaveEpisodeProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		episodeID       int64
		positionSeconds int
		setupMock       func(repo *mocks.MockMovieRepo)
		wantErr         error
	}{
		{
			name:      "invalid episode id",
			episodeID: 0,
			wantErr:   domain.ErrInvalidEpisodeID,
		},
		{
			name:            "invalid position",
			episodeID:       10,
			positionSeconds: -1,
			wantErr:         domain.ErrInvalidWatchProgress,
		},
		{
			name:            "repo error",
			episodeID:       10,
			positionSeconds: 5,
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					UpsertWatchProgress(gomock.Any(), int64(13), int64(10), 5).
					Return(errors.New("save failed"))
			},
			wantErr: errors.New("save failed"),
		},
		{
			name:            "success",
			episodeID:       10,
			positionSeconds: 5,
			setupMock: func(repo *mocks.MockMovieRepo) {
				repo.EXPECT().
					UpsertWatchProgress(gomock.Any(), int64(13), int64(10), 5).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMovieRepo(ctrl)
			actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			u := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)
			err := u.SaveEpisodeProgress(context.Background(), 13, tt.episodeID, tt.positionSeconds)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) && (err == nil || err.Error() != tt.wantErr.Error()) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMovieUsecase_Search(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		query     string
		setupMock func(repo *mocks.MockMovieRepo, actorStorage, cardStorage *mocks.MockFileStorage)
		wantErr   error
		want      domain.SearchResponse
	}{
		{
			name:    "invalid query",
			query:   "   ",
			wantErr: domain.ErrInvalidSearchQuery,
		},
		{
			name:  "repo error",
			query: "dune",
			setupMock: func(repo *mocks.MockMovieRepo, actorStorage, cardStorage *mocks.MockFileStorage) {
				repo.EXPECT().
					Search(gomock.Any(), "dune").
					Return(domain.SearchResponse{}, errors.New("search failed"))
			},
			wantErr: errors.New("search failed"),
		},
		{
			name:  "success",
			query: "dune",
			setupMock: func(repo *mocks.MockMovieRepo, actorStorage, cardStorage *mocks.MockFileStorage) {
				repo.EXPECT().
					Search(gomock.Any(), "dune").
					Return(domain.SearchResponse{
						Query:  "dune",
						Movies: []domain.MoviePreview{{ID: 1, Title: "Dune", ImgUrl: "cards/dune.jpg"}},
						Actors: []domain.ActorPreview{{ID: 2, FullName: "Zendaya", PictureFileKey: "actors/zendaya.jpg"}},
					}, nil)
				cardStorage.EXPECT().
					PresignGetObject(gomock.Any(), "cards/dune.jpg", time.Duration(0)).
					Return("https://cdn.example/cards/dune.jpg", nil)
				actorStorage.EXPECT().
					PresignGetObject(gomock.Any(), "actors/zendaya.jpg", time.Duration(0)).
					Return("https://cdn.example/actors/zendaya.jpg", nil)
			},
			want: domain.SearchResponse{
				Query:  "dune",
				Movies: []domain.MoviePreview{{ID: 1, Title: "Dune", ImgUrl: "https://cdn.example/cards/dune.jpg"}},
				Actors: []domain.ActorPreview{{ID: 2, FullName: "Zendaya", PictureFileKey: "https://cdn.example/actors/zendaya.jpg"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockMovieRepo(ctrl)
			actorStorage, posterStorage, cardStorage, videoStorage := newMovieStorages(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(repo, actorStorage, cardStorage)
			}

			u := newMovieUsecase(repo, actorStorage, posterStorage, cardStorage, videoStorage)
			got, err := u.Search(context.Background(), tt.query)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) && (err == nil || err.Error() != tt.wantErr.Error()) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Query != tt.want.Query || got.Movies[0].ImgUrl != tt.want.Movies[0].ImgUrl ||
				got.Actors[0].PictureFileKey != tt.want.Actors[0].PictureFileKey {
				t.Fatalf("expected search result %#v, got %#v", tt.want, got)
			}
		})
	}
}
