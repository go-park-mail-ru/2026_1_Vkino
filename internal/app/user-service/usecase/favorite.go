package usecase

import (
	"context"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
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

func (u *UserUsecase) ToggleFavorite(
	ctx context.Context,
	userID, movieID int64,
) (domain.FavoriteMovieResponse, error) {
	if movieID <= 0 {
		return domain.FavoriteMovieResponse{}, domain.ErrInvalidMovieID
	}
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.FavoriteMovieResponse{}, domain.ErrInvalidToken
	}

	isFavorite, err := u.userRepo.ToggleFavorite(ctx, userID, movieID)
	if err != nil {
		return domain.FavoriteMovieResponse{}, domain.ErrInternal
	}

	return domain.FavoriteMovieResponse{
		MovieID:    movieID,
		IsFavorite: isFavorite,
	}, nil
}

func (u *UserUsecase) GetFavorites(
	ctx context.Context,
	userID int64,
	limit, offset int32,
) (domain.FavoritesResponse, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.FavoritesResponse{}, domain.ErrInvalidToken
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	movieIDs, total, err := u.userRepo.GetFavorites(ctx, userID, limit, offset)
	if err != nil {
		return domain.FavoritesResponse{}, domain.ErrInternal
	}

	return domain.FavoritesResponse{
		MovieIDs:   movieIDs,
		TotalCount: total,
	}, nil
}
