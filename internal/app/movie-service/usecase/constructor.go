package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/repository"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/subscription"
)

func NewMovieUsecase(
	movieRepo repository.MovieRepo,
	subscriptionReader subscription.StateReader,
	posterStore storage.FileStorage,
	cardStore storage.FileStorage,
	actorStore storage.FileStorage,
	videoStore storage.FileStorage,
) *MovieUsecase {
	return &MovieUsecase{
		movieRepo:          movieRepo,
		subscriptionReader: subscriptionReader,
		posterStore:        posterStore,
		cardStore:          cardStore,
		actorStore:         actorStore,
		videoStore:         videoStore,
	}
}
