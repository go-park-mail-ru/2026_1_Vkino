package usecase

import (
	"context"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

const (
	minMovieRating = 0
	maxMovieRating = 10
)

func (u *UserUsecase) SetMovieRating(
	ctx context.Context,
	userID, movieID int64,
	rating float64,
) (domain.MovieRatingResponse, error) {
	if movieID <= 0 {
		return domain.MovieRatingResponse{}, domain.ErrInvalidMovieID
	}

	if rating < minMovieRating || rating > maxMovieRating {
		return domain.MovieRatingResponse{}, domain.ErrInvalidMovieRating
	}

	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.MovieRatingResponse{}, domain.ErrInvalidToken
	}

	if err := u.userRepo.SetMovieRating(ctx, userID, movieID, rating); err != nil {
		return domain.MovieRatingResponse{}, err
	}

	return domain.MovieRatingResponse{
		MovieID: movieID,
		Rating:  rating,
	}, nil
}
