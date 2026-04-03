package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

type Usecase interface {
	GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error)
	GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error)
	GetMovieByID(ctx context.Context, id int64) (domain.MovieResponse, error)
	GetActorByID(ctx context.Context, id int64) (domain.ActorResponse, error)
}
