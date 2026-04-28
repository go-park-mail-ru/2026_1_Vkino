package usecase

import (
	"context"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

func (u *MovieUsecase) GetMovieByID(ctx context.Context, movieID int64) (domain.MovieResponse, error) {
	if movieID <= 0 {
		return domain.MovieResponse{}, domain.ErrInvalidMovieID
	}

	movie, err := u.movieRepo.GetMovieByID(ctx, movieID)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	resp, err := u.buildMovieResponse(ctx, movie)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	if authCtx, err := authctx.FromContext(ctx); err == nil {
		isFavorite, err := u.movieRepo.IsFavorite(ctx, authCtx.UserID, movieID)
		if err == nil {
			resp.IsFavorite = isFavorite
		}
	}

	return resp, nil
}

func (u *MovieUsecase) GetMoviesByIDs(ctx context.Context, movieIDs []int64) ([]domain.MovieCardResponse, error) {
	movies, err := u.movieRepo.GetMovieCardsByIDs(ctx, movieIDs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.MovieCardResponse, 0, len(movies))
	for _, movie := range movies {
		card, err := u.buildMovieCardResponse(ctx, movie)
		if err != nil {
			return nil, err
		}

		result = append(result, card)
	}

	return result, nil
}
