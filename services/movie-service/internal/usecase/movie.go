package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/domain"
)

func (u *MovieUsecase) GetMovieByID(ctx context.Context, movieID int64) (domain.MovieResponse, error) {
	if movieID <= 0 {
		return domain.MovieResponse{}, domain.ErrInvalidMovieID
	}

	movie, err := u.movieRepo.GetMovieByID(ctx, movieID)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	return u.buildMovieResponse(ctx, movie)
}
