package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

func (m *MovieUsecase) GetMovieByID(ctx context.Context, id int64) (domain.MovieResponse, error) {
	if id <= 0 {
		return domain.MovieResponse{}, domain.ErrInvalidMovieID
	}

	movie, err := m.movieRepo.GetMovieByID(ctx, id)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	movie.PictureFileKey, err = m.presignCardURL(ctx, movie.PictureFileKey)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	movie.PosterFileKey, err = m.presignPosterURL(ctx, movie.PosterFileKey)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	for i := range movie.Actors {
		movie.Actors[i].PictureFileKey, err = m.presignActorURL(ctx, movie.Actors[i].PictureFileKey)
		if err != nil {
			return domain.MovieResponse{}, err
		}
	}

	episodes, err := m.movieRepo.GetEpisodesByMovieID(ctx, id)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	movie.Episodes = episodes

	return movie, nil
}
