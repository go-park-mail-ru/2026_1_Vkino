package usecase

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/repository/mocks"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"go.uber.org/mock/gomock"
)

func TestGetContinueWatching_CallsRepoWithUserAndLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	mr := mocks.NewMockMovieRepo(ctrl)
	u := NewMovieUsecase(mr, smartContinueStateReader{}, nil, nil, nil, nil)

	mr.EXPECT().GetContinueWatching(gomock.Any(), int64(42), int32(7)).Return(nil, nil)

	_, err := u.GetContinueWatching(context.Background(), 42, 7)
	if err != nil {
		t.Fatalf("GetContinueWatching: %v", err)
	}
}

func TestGetWatchHistory_CallsRepoWithMinProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	mr := mocks.NewMockMovieRepo(ctrl)
	u := NewMovieUsecase(mr, smartContinueStateReader{}, nil, nil, nil, nil)

	mr.EXPECT().GetWatchHistory(gomock.Any(), int64(3), int32(10), 0.95).Return(nil, nil)

	_, err := u.GetWatchHistory(context.Background(), 3, 10, 0.95)
	if err != nil {
		t.Fatalf("GetWatchHistory: %v", err)
	}
}

func TestGetContinueWatching_PresignsPosterURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	mr := mocks.NewMockMovieRepo(ctrl)
	u := NewMovieUsecase(mr, smartContinueStateReader{}, stubFileStorage{}, nil, nil, nil)

	mr.EXPECT().GetContinueWatching(gomock.Any(), int64(42), int32(5)).Return([]domain.WatchProgressItem{
		{
			MovieID:    10,
			MovieTitle: "Movie",
			PosterURL:  "posters/movie.webp",
		},
	}, nil)

	items, err := u.GetContinueWatching(context.Background(), 42, 5)
	if err != nil {
		t.Fatalf("GetContinueWatching: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}

	if got := items[0].PosterURL; got != "https://cdn.test/posters/movie.webp" {
		t.Fatalf("poster_url = %q, want %q", got, "https://cdn.test/posters/movie.webp")
	}
}

type smartContinueStateReader struct{}

func (smartContinueStateReader) GetSubscriptionState(
	context.Context,
	int64,
) (*userv1.GetSubscriptionCapabilitiesResponse, error) {
	state := defaultSubscriptionState()
	state.Capabilities.CanUseSmartContinue = true

	return state, nil
}
