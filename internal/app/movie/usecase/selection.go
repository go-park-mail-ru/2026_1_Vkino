package usecase

import (
	"context"

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

func (m *MovieUsecase) GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error) {
	selections, err := m.movieRepo.GetAllSelections(ctx)

	if err != nil {
		return nil, err
	}

	return selections, nil
}

func (m *MovieUsecase) GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error) {
	selection, err := m.movieRepo.GetSelectionByTitle(ctx, title)

	if err != nil {
		return domain.SelectionResponse{}, err
	}

	return selection, nil
}
