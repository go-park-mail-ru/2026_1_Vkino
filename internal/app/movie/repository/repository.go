package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

type MovieRepo interface {
	GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error)
	GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error)
	GetMovieByID(ctx context.Context, id int64) (domain.MovieResponse, error)
	GetActorByID(ctx context.Context, id int64) (domain.ActorResponse, error)

	GetEpisodesByMovieID(ctx context.Context, movieID int64) ([]domain.EpisodeItemResponse, error)
	GetEpisodePlayback(ctx context.Context, episodeID int64) (domain.EpisodePlaybackResponse, error)
	GetWatchProgress(ctx context.Context, userID, episodeID int64) (int, error)
	UpsertWatchProgress(ctx context.Context, userID, episodeID int64, positionSeconds int) error
}
