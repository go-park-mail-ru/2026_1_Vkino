package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/movie-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/movie-service/repository"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

type Usecase interface {
	GetMovieByID(ctx context.Context, movieID int64) (domain.MovieResponse, error)
	GetActorByID(ctx context.Context, actorID int64) (domain.ActorResponse, error)
	GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error)
	GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error)

	SearchMovies(ctx context.Context, query string) ([]domain.MovieCardResponse, error)
	GetEpisodePlayback(ctx context.Context, episodeID int64) (domain.EpisodePlaybackResponse, error)
	GetEpisodeProgress(ctx context.Context, userID, episodeID int64) (domain.EpisodeProgressResponse, error)
	SaveEpisodeProgress(ctx context.Context, userID, episodeID, positionSec int64) (domain.EpisodeProgressResponse, error)
}

type MovieUsecase struct {
	movieRepo   repository.MovieRepo
	posterStore storage.FileStorage
	cardStore   storage.FileStorage
	actorStore  storage.FileStorage
	videoStore  storage.FileStorage
}
