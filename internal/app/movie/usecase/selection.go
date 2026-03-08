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

func (m *MovieUsecase) GetAllSelections() ([]domain.SelectionResponse, error) {
	selections, err := m.movieRepo.GetAllSelections()

	if err != nil {
		return nil, err
	}
	return selections, nil
}

func (m *MovieUsecase) GetSelectionByTitle(title string) (*domain.SelectionResponse, error) {
	selection, err := m.movieRepo.GetSelectionByTitle(title)

	if err != nil {
		return nil, err
	}

	return selection, nil
}
