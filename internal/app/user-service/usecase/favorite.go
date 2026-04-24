package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) AddMovieToFavorites(
	ctx context.Context,
	userID, movieID int64,
) (domain2.FavoriteMovieResponse, error) {
	if movieID <= 0 {
		return domain2.FavoriteMovieResponse{}, domain2.ErrInvalidMovieID
	}

	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain2.FavoriteMovieResponse{}, domain2.ErrInvalidToken
	}

	if err := u.userRepo.AddMovieToFavorites(ctx, userID, movieID); err != nil {
		return domain2.FavoriteMovieResponse{}, err
	}

	return domain2.FavoriteMovieResponse{
		MovieID:    movieID,
		IsFavorite: true,
	}, nil
}
