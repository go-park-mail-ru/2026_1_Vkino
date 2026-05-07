package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/mocks"
	"go.uber.org/mock/gomock"
)

func TestSetMovieRating_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)
	usecase := NewUserUsecase(repo, nil, nil)

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		SetMovieRating(gomock.Any(), int64(42), int64(7), 8.5).
		Return(nil)

	resp, err := usecase.SetMovieRating(context.Background(), 42, 7, 8.5)
	if err != nil {
		t.Fatalf("SetMovieRating() error = %v", err)
	}

	if resp.MovieID != 7 || resp.Rating != 8.5 {
		t.Fatalf("SetMovieRating() = %+v, want movie_id=7 rating=8.5", resp)
	}
}

func TestSetMovieRating_InvalidRating(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)
	usecase := NewUserUsecase(repo, nil, nil)

	_, err := usecase.SetMovieRating(context.Background(), 42, 7, 10.5)
	if !errors.Is(err, domain.ErrInvalidMovieRating) {
		t.Fatalf("SetMovieRating() error = %v, want %v", err, domain.ErrInvalidMovieRating)
	}
}

func TestSetMovieRating_InvalidToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)
	usecase := NewUserUsecase(repo, nil, nil)

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(nil, domain.ErrUserNotFound)

	_, err := usecase.SetMovieRating(context.Background(), 42, 7, 8)
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("SetMovieRating() error = %v, want %v", err, domain.ErrInvalidToken)
	}
}
