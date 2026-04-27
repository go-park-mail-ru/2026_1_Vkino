package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) GetMovieByID(ctx context.Context, movieID int64, userID int64) (domain2.MovieResponse, error) {
	if movieID <= 0 {
		return domain2.MovieResponse{}, domain2.ErrInvalidMovieID
	}

	movie, err := u.movieRepo.GetMovieByID(ctx, movieID)
	if err != nil {
		return domain2.MovieResponse{}, err
	}

	resp, err := u.buildMovieResponse(ctx, movie)
	if err != nil {
		return domain2.MovieResponse{}, err
	}

	if userID > 0 {
		isFavorite, err := u.movieRepo.IsFavorite(ctx, userID, movieID)
		if err == nil {
			resp.IsFavorite = &isFavorite
		}
	}

	return resp, nil
}
