package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/domain"
)

type MovieRepo interface {
	GetMovieByID(ctx context.Context, movieID int64) (*domain.Movie, error)
	GetActorByID(ctx context.Context, actorID int64) (*domain.Actor, error)
	GetSelectionByTitle(ctx context.Context, title string) (domain.Selection, error)
	GetAllSelections(ctx context.Context) ([]domain.Selection, error)
}
