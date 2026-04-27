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

func (u *UserUsecase) ToggleFavorite(
	ctx context.Context,
	userID, movieID int64,
) (domain2.FavoriteMovieResponse, error) {
	if movieID <= 0 {
		return domain2.FavoriteMovieResponse{}, domain2.ErrInvalidMovieID
	}
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain2.FavoriteMovieResponse{}, domain2.ErrInvalidToken
	}

	isFavorite, err := u.userRepo.ToggleFavorite(ctx, userID, movieID)
	if err != nil {
		return domain2.FavoriteMovieResponse{}, domain2.ErrInternal
	}

	return domain2.FavoriteMovieResponse{
		MovieID:    movieID,
		IsFavorite: isFavorite,
	}, nil
}

func (u *UserUsecase) GetFavorites(
	ctx context.Context,
	userID int64,
	limit, offset int32,
) (domain2.FavoritesResponse, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain2.FavoritesResponse{}, domain2.ErrInvalidToken
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	movies, total, err := u.userRepo.GetFavorites(ctx, userID, limit, offset)
	if err != nil {
		return domain2.FavoritesResponse{}, domain2.ErrInternal
	}

	return domain2.FavoritesResponse{
		Movies:     movies,
		TotalCount: total,
	}, nil
}
