package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/user-service/domain"
)

func (u *UserUsecase) AddMovieToFavorites(
	ctx context.Context,
	userID, movieID int64,
) (domain.FavoriteMovieResponse, error) {
	if movieID <= 0 {
		return domain.FavoriteMovieResponse{}, domain.ErrInvalidMovieID
	}

	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.FavoriteMovieResponse{}, domain.ErrInvalidToken
	}

	if err := u.userRepo.AddMovieToFavorites(ctx, userID, movieID); err != nil {
		return domain.FavoriteMovieResponse{}, err
	}

	return domain.FavoriteMovieResponse{
		MovieID:    movieID,
		IsFavorite: true,
	}, nil
}
