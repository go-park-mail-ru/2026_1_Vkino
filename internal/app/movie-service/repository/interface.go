package repository

//go:generate mockgen -source=./interface.go -destination=./mocks/movie_repo_mock.go -package=mocks MovieRepo

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

type MovieRepo interface {
	GetMovieByID(ctx context.Context, movieID int64) (*domain.Movie, error)
	GetActorByID(ctx context.Context, actorID int64) (*domain.Actor, error)
	GetGenreByID(ctx context.Context, genreID int64) (domain.Genre, error)
	GetAllGenres(ctx context.Context) ([]domain.GenreShort, error)
	GetSelectionByTitle(ctx context.Context, title string) (domain.Selection, error)
	GetAllSelections(ctx context.Context) ([]domain.Selection, error)
	GetMovieCardsByIDs(ctx context.Context, movieIDs []int64) ([]domain.MovieCard, error)

	SearchMovies(ctx context.Context, query string) ([]domain.MovieCard, error)
	GetEpisodePlayback(ctx context.Context, episodeID int64) (*domain.Episode, error)
	GetEpisodeProgress(ctx context.Context, userID, episodeID int64) (domain.EpisodeProgress, error)
	SaveEpisodeProgress(ctx context.Context, userID, episodeID, positionSec int64) (domain.EpisodeProgress, error)
	IsFavorite(ctx context.Context, userID, movieID int64) (bool, error)
	GetContinueWatching(ctx context.Context, userID int64, limit int32) ([]domain.WatchProgressItem, error)
	GetWatchHistory(ctx context.Context, userID int64, limit int32, minProgress float64) ([]domain.WatchProgressItem, error)
}
