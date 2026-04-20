package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/repository"
)

type Usecase interface {
	GetMovieByID(ctx context.Context, movieID int64) (domain.MovieResponse, error)
	GetActorByID(ctx context.Context, actorID int64) (domain.ActorResponse, error)
	GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error)
	GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error)
}

type MovieUsecase struct {
	movieRepo   repository.MovieRepo
	posterStore storage.FileStorage
	cardStore   storage.FileStorage
	actorStore  storage.FileStorage
}
