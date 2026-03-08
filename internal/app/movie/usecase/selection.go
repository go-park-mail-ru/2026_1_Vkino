package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/repository"
)

type MovieUsecase struct {
	movieRepo repository.MovieRepo
}

func NewMovieUsecase(movieRepo repository.MovieRepo) *MovieUsecase {
	return &MovieUsecase{
		movieRepo: movieRepo,
	}
}

func (m *MovieUsecase) GetSelectionByTitle(title string) (*domain.SelectionResponse, error) {
	selections, err := m.movieRepo.GetSelectionByTitle(title)

	if err != nil {
		return nil, err
	}

	return selections, nil
}
