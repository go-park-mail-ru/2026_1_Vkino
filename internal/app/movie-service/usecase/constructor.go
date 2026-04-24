package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/repository"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

func NewMovieUsecase(
	movieRepo repository.MovieRepo,
	posterStore storage.FileStorage,
	cardStore storage.FileStorage,
	actorStore storage.FileStorage,
	videoStore storage.FileStorage,
) *MovieUsecase {
	return &MovieUsecase{
		movieRepo:   movieRepo,
		posterStore: posterStore,
		cardStore:   cardStore,
		actorStore:  actorStore,
		videoStore:  videoStore,
	}
}
