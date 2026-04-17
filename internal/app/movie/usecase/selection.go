package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/repository"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

type MovieUsecase struct {
	movieRepo     repository.MovieRepo
	actorStorage  storage.FileStorage
	posterStorage storage.FileStorage
	cardStorage   storage.FileStorage
	videoStorage  storage.FileStorage
}

func NewMovieUsecase(
	movieRepo repository.MovieRepo,
	actorStorage storage.FileStorage,
	posterStorage storage.FileStorage,
	cardStorage storage.FileStorage,
	videoStorage storage.FileStorage,
) *MovieUsecase {
	return &MovieUsecase{
		movieRepo:     movieRepo,
		actorStorage:  actorStorage,
		posterStorage: posterStorage,
		cardStorage:   cardStorage,
		videoStorage:  videoStorage,
	}
}

func (m *MovieUsecase) GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error) {
	selections, err := m.movieRepo.GetAllSelections(ctx)
	if err != nil {
		return nil, err
	}

	for i := range selections {
		for j := range selections[i].Movies {
			selections[i].Movies[j].ImgUrl, err = m.presignCardURL(ctx, selections[i].Movies[j].ImgUrl)
			if err != nil {
				return nil, err
			}
		}
	}

	return selections, nil
}

func (m *MovieUsecase) GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error) {
	selection, err := m.movieRepo.GetSelectionByTitle(ctx, title)
	if err != nil {
		return domain.SelectionResponse{}, err
	}

	for i := range selection.Movies {
		selection.Movies[i].ImgUrl, err = m.presignCardURL(ctx, selection.Movies[i].ImgUrl)
		if err != nil {
			return domain.SelectionResponse{}, err
		}
	}

	return selection, nil
}
