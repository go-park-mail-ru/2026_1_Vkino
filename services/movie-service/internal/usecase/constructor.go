package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/repository"
)

func NewMovieUsecase(
	movieRepo repository.MovieRepo,
	posterStore storage.FileStorage,
	cardStore storage.FileStorage,
	actorStore storage.FileStorage,
) *MovieUsecase {
	return &MovieUsecase{
		movieRepo:   movieRepo,
		posterStore: posterStore,
		cardStore:   cardStore,
		actorStore:  actorStore,
	}
}
