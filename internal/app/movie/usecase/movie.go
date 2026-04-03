package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

func (m *MovieUsecase) GetMovieByID(ctx context.Context, id int64) (domain.MovieResponse, error) {
	movie, err := m.movieRepo.GetMovieByID(ctx, id)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	return movie, nil
}
