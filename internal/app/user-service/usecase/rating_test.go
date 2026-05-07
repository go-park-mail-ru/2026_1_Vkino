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
	rating := 8.5

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		SetMovieReview(gomock.Any(), int64(42), int64(7), &rating, (*string)(nil)).
		Return(domain.MovieReviewResponse{
			ReviewID: 1,
			MovieID:  7,
			Rating:   &rating,
		}, nil)

	resp, err := usecase.SetMovieRating(context.Background(), 42, 7, 8.5)
	if err != nil {
		t.Fatalf("SetMovieRating() error = %v", err)
	}

	if resp.MovieID != 7 || resp.Rating != 8.5 {
		t.Fatalf("SetMovieRating() = %+v, want movie_id=7 rating=8.5", resp)
	}
}

func TestSetMovieReview_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)
	usecase := NewUserUsecase(repo, nil, nil)

	rating := 7.5
	comment := "Worth watching"

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		SetMovieReview(gomock.Any(), int64(42), int64(7), &rating, &comment).
		Return(domain.MovieReviewResponse{
			ReviewID: 12,
			MovieID:  7,
			Rating:   &rating,
			Comment:  &comment,
		}, nil)

	resp, err := usecase.SetMovieReview(context.Background(), 42, 7, &rating, &comment)
	if err != nil {
		t.Fatalf("SetMovieReview() error = %v", err)
	}

	if resp.ReviewID != 12 || resp.Comment == nil || *resp.Comment != comment {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestSetMovieReview_InvalidPayload(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)
	usecase := NewUserUsecase(repo, nil, nil)

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{ID: 42}, nil)

	_, err := usecase.SetMovieReview(context.Background(), 42, 7, nil, nil)
	if !errors.Is(err, domain.ErrInvalidMovieReview) {
		t.Fatalf("SetMovieReview() error = %v, want %v", err, domain.ErrInvalidMovieReview)
	}
}

func TestSetReviewReaction_InvalidVote(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)
	usecase := NewUserUsecase(repo, nil, nil)

	_, err := usecase.SetReviewReaction(context.Background(), 42, 10, "heart")
	if !errors.Is(err, domain.ErrInvalidReviewVote) {
		t.Fatalf("SetReviewReaction() error = %v, want %v", err, domain.ErrInvalidReviewVote)
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
