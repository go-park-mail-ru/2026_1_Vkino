package usecase

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/repository/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetContinueWatching_CallsRepoWithUserAndLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	mr := mocks.NewMockMovieRepo(ctrl)
	u := NewMovieUsecase(mr, nil, nil, nil, nil)

	mr.EXPECT().GetContinueWatching(gomock.Any(), int64(42), int32(7)).Return(nil, nil)

	_, err := u.GetContinueWatching(context.Background(), 42, 7)
	if err != nil {
		t.Fatalf("GetContinueWatching: %v", err)
	}
}

func TestGetWatchHistory_CallsRepoWithMinProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	mr := mocks.NewMockMovieRepo(ctrl)
	u := NewMovieUsecase(mr, nil, nil, nil, nil)

	mr.EXPECT().GetWatchHistory(gomock.Any(), int64(3), int32(10), 0.95).Return(nil, nil)

	_, err := u.GetWatchHistory(context.Background(), 3, 10, 0.95)
	if err != nil {
		t.Fatalf("GetWatchHistory: %v", err)
	}
}
