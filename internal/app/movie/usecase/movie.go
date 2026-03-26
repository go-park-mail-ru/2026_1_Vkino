package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/google/uuid"
)

func (m *MovieUsecase) GetMovieByID(id uuid.UUID) (domain.MovieResponse, error) {
	movie, err := m.movieRepo.GetMovieByID(id)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	return movie, nil
}