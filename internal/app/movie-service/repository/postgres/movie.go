//nolint:lll,wsl_v5 // Repository methods are kept close to SQL/query parameters for readability.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/jackc/pgx/v5"
)

type MovieRepo struct {
	db *corepostgres.Client
}

func NewMovieRepo(db *corepostgres.Client) *MovieRepo {
	return &MovieRepo{db: db}
}

var (
	ErrMovieNotFound     = errors.New("movie not found")
	ErrActorNotFound     = errors.New("actor not found")
	ErrGenreNotFound     = errors.New("genre not found")
	ErrSelectionNotFound = errors.New("selection not found")
	ErrEpisodeNotFound   = errors.New("episode not found")
)

func (r *MovieRepo) GetMovieByID(ctx context.Context, movieID int64) (*domain.Movie, error) {
	var movie domain.Movie

	err := r.db.QueryRow(ctx, sqlGetMovieBaseByID, movieID).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.Director,
		&movie.TrailerURL,
		&movie.ContentType,
		&movie.ReleaseYear,
		&movie.DurationSeconds,
		&movie.AgeLimit,
		&movie.OriginalLanguageID,
		&movie.OriginalLanguage,
		&movie.CountryID,
		&movie.Country,
		&movie.PictureFileKey,
		&movie.PosterFileKey,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMovieNotFound
		}

		return nil, fmt.Errorf("get movie by id: %w", err)
	}

	movie.Genres, err = r.getMovieGenres(ctx, movieID)
	if err != nil {
		return nil, err
	}

	movie.Actors, err = r.getMovieActors(ctx, movieID)
	if err != nil {
		return nil, err
	}

	movie.Episodes, err = r.getMovieEpisodes(ctx, movieID)
	if err != nil {
		return nil, err
	}

	return &movie, nil
}

func (r *MovieRepo) GetActorByID(ctx context.Context, actorID int64) (*domain.Actor, error) {
	var actor domain.Actor

	err := r.db.QueryRow(ctx, sqlGetActorBaseByID, actorID).Scan(
		&actor.ID,
		&actor.FullName,
		&actor.BirthDate,
		&actor.Biography,
		&actor.CountryID,
		&actor.PictureFileKey,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrActorNotFound
		}

		return nil, fmt.Errorf("get actor by id: %w", err)
	}

	actor.Movies, err = r.getActorMovies(ctx, actorID)
	if err != nil {
		return nil, err
	}

	return &actor, nil
}

func (r *MovieRepo) GetGenreByID(ctx context.Context, genreID int64) (domain.Genre, error) {
	var genre domain.Genre

	err := r.db.QueryRow(ctx, sqlGetGenreBaseByID, genreID).Scan(
		&genre.ID,
		&genre.Title,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Genre{}, ErrGenreNotFound
		}

		return domain.Genre{}, fmt.Errorf("get genre by id: %w", err)
	}

	genre.Movies, err = r.getGenreMovies(ctx, genreID)
	if err != nil {
		return domain.Genre{}, err
	}

	return genre, nil
}

func (r *MovieRepo) GetAllGenres(ctx context.Context) ([]domain.GenreShort, error) {
	rows, err := r.db.Query(ctx, sqlGetAllGenres)
	if err != nil {
		return nil, fmt.Errorf("get all genres: %w", err)
	}
	defer rows.Close()

	genres := make([]domain.GenreShort, 0)
	for rows.Next() {
		var genre domain.GenreShort
		if err = rows.Scan(&genre.ID, &genre.Title); err != nil {
			return nil, fmt.Errorf("scan genre: %w", err)
		}

		genres = append(genres, genre)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate genres: %w", err)
	}

	return genres, nil
}

func (r *MovieRepo) GetSelectionByTitle(ctx context.Context, title string) (domain.Selection, error) {
	rows, err := r.db.Query(ctx, sqlGetSelectionMoviesByTitle, title)
	if err != nil {
		return domain.Selection{}, fmt.Errorf("get selection by title: %w", err)
	}
	defer rows.Close()

	selection := domain.Selection{
		Movies: make([]domain.MovieCard, 0),
	}

	for rows.Next() {
		var movie domain.MovieCard
		if err = rows.Scan(
			&selection.Title,
			&movie.ID,
			&movie.Title,
			&movie.PictureFileKey,
		); err != nil {
			return domain.Selection{}, fmt.Errorf("scan selection movie: %w", err)
		}

		selection.Movies = append(selection.Movies, movie)
	}

	if err = rows.Err(); err != nil {
		return domain.Selection{}, fmt.Errorf("iterate selection rows: %w", err)
	}

	if selection.Title == "" {
		return domain.Selection{}, ErrSelectionNotFound
	}

	return selection, nil
}

func (r *MovieRepo) GetAllSelections(ctx context.Context) ([]domain.Selection, error) {
	rows, err := r.db.Query(ctx, sqlGetAllSelectionMovies)
	if err != nil {
		return nil, fmt.Errorf("get all selections: %w", err)
	}
	defer rows.Close()

	selectionMap := make(map[string]*domain.Selection)
	order := make([]string, 0)

	for rows.Next() {
		var (
			title string
			movie domain.MovieCard
		)

		if err = rows.Scan(
			&title,
			&movie.ID,
			&movie.Title,
			&movie.PictureFileKey,
		); err != nil {
			return nil, fmt.Errorf("scan all selections row: %w", err)
		}

		selection, ok := selectionMap[title]
		if !ok {
			selection = &domain.Selection{
				Title:  title,
				Movies: make([]domain.MovieCard, 0),
			}
			selectionMap[title] = selection
			order = append(order, title)
		}

		selection.Movies = append(selection.Movies, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate all selections rows: %w", err)
	}

	result := make([]domain.Selection, 0, len(order))
	for _, title := range order {
		result = append(result, *selectionMap[title])
	}

	return result, nil
}

func (r *MovieRepo) GetMovieCardsByIDs(ctx context.Context, movieIDs []int64) ([]domain.MovieCard, error) {
	if len(movieIDs) == 0 {
		return []domain.MovieCard{}, nil
	}

	rows, err := r.db.Query(ctx, sqlGetMovieCardsByIDs, movieIDs)
	if err != nil {
		return nil, fmt.Errorf("get movie cards by ids: %w", err)
	}
	defer rows.Close()

	result := make([]domain.MovieCard, 0, len(movieIDs))

	for rows.Next() {
		var movie domain.MovieCard
		if err = rows.Scan(&movie.ID, &movie.Title, &movie.PictureFileKey); err != nil {
			return nil, fmt.Errorf("scan movie card: %w", err)
		}

		result = append(result, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate movie cards: %w", err)
	}

	return result, nil
}

func (r *MovieRepo) SearchMovies(ctx context.Context, query string) ([]domain.MovieCard, error) {
	rows, err := r.db.Query(ctx, sqlSearchMovies, query)
	if err != nil {
		return nil, fmt.Errorf("search movies: %w", err)
	}
	defer rows.Close()

	result := make([]domain.MovieCard, 0)

	for rows.Next() {
		var movie domain.MovieCard
		if err = rows.Scan(&movie.ID, &movie.Title, &movie.PictureFileKey); err != nil {
			return nil, fmt.Errorf("scan searched movie: %w", err)
		}

		result = append(result, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate searched movies: %w", err)
	}

	return result, nil
}

func (r *MovieRepo) SearchActors(ctx context.Context, query string) ([]domain.ActorShort, error) {
	rows, err := r.db.Query(ctx, sqlSearchActors, query)
	if err != nil {
		return nil, fmt.Errorf("search actors: %w", err)
	}
	defer rows.Close()

	result := make([]domain.ActorShort, 0)

	for rows.Next() {
		var actor domain.ActorShort
		if err = rows.Scan(&actor.ID, &actor.FullName, &actor.PictureFileKey); err != nil {
			return nil, fmt.Errorf("scan searched actor: %w", err)
		}

		result = append(result, actor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate searched actors: %w", err)
	}

	return result, nil
}

func (r *MovieRepo) GetEpisodePlayback(ctx context.Context, episodeID int64) (*domain.Episode, error) {
	var episode domain.Episode

	err := r.db.QueryRow(ctx, sqlGetEpisodePlayback, episodeID).Scan(
		&episode.ID,
		&episode.MovieID,
		&episode.SeasonNumber,
		&episode.EpisodeNumber,
		&episode.Title,
		&episode.DurationSeconds,
		&episode.VideoFileKey,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEpisodeNotFound
		}

		return nil, fmt.Errorf("get episode playback: %w", err)
	}

	return &episode, nil
}

func (r *MovieRepo) GetEpisodeProgress(ctx context.Context, userID, episodeID int64) (domain.EpisodeProgress, error) {
	var progress domain.EpisodeProgress

	err := r.db.QueryRow(ctx, sqlGetEpisodeProgress, userID, episodeID).Scan(
		&progress.EpisodeID,
		&progress.PositionSeconds,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.EpisodeProgress{
				EpisodeID:       episodeID,
				PositionSeconds: 0,
			}, nil
		}

		return domain.EpisodeProgress{}, fmt.Errorf("get episode progress: %w", err)
	}

	return progress, nil
}

func (r *MovieRepo) SaveEpisodeProgress(
	ctx context.Context,
	userID, episodeID, positionSec int64,
) (domain.EpisodeProgress, error) {
	var progress domain.EpisodeProgress

	err := r.db.QueryRow(ctx, sqlSaveEpisodeProgress, userID, episodeID, positionSec).Scan(
		&progress.EpisodeID,
		&progress.PositionSeconds,
	)
	if err != nil {
		return domain.EpisodeProgress{}, fmt.Errorf("save episode progress: %w", err)
	}

	return progress, nil
}

func (r *MovieRepo) IsFavorite(ctx context.Context, userID, movieID int64) (bool, error) {
	var isFavorite bool
	if err := r.db.QueryRow(ctx, sqlIsFavorite, userID, movieID).Scan(&isFavorite); err != nil {
		return false, fmt.Errorf("is favorite: %w", err)
	}

	return isFavorite, nil
}

func (r *MovieRepo) GetContinueWatching(ctx context.Context, userID int64, limit int32) ([]domain.WatchProgressItem, error) {
	return r.getWatchProgressItems(ctx, sqlGetContinueWatching, userID, limit, 0)
}

func (r *MovieRepo) GetWatchHistory(ctx context.Context, userID int64, limit int32, minProgress float64) ([]domain.WatchProgressItem, error) {
	return r.getWatchProgressItems(ctx, sqlGetWatchHistory, userID, limit, minProgress)
}

func (r *MovieRepo) getWatchProgressItems(ctx context.Context, query string, userID int64, limit int32, minProgress float64) ([]domain.WatchProgressItem, error) {
	rows, err := r.db.Query(ctx, query, userID, limit, minProgress)
	if err != nil {
		return nil, fmt.Errorf("get watch progress: %w", err)
	}
	defer rows.Close()

	items := make([]domain.WatchProgressItem, 0, limit)

	for rows.Next() {
		var (
			item      domain.WatchProgressItem
			updatedAt time.Time
		)
		if err := rows.Scan(
			&item.EpisodeID,
			&item.MovieID,
			&item.MovieTitle,
			&item.PosterURL,
			&item.ContentType,
			&item.SeasonNumber,
			&item.EpisodeNumber,
			&item.EpisodeTitle,
			&item.PositionSeconds,
			&item.DurationSeconds,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan watch progress item: %w", err)
		}

		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate watch progress: %w", err)
	}

	return items, nil
}

func (r *MovieRepo) getMovieGenres(ctx context.Context, movieID int64) ([]string, error) {
	rows, err := r.db.Query(ctx, sqlGetMovieGenresByID, movieID)
	if err != nil {
		return nil, fmt.Errorf("get movie genres: %w", err)
	}
	defer rows.Close()

	result := make([]string, 0)

	for rows.Next() {
		var item string
		if err = rows.Scan(&item); err != nil {
			return nil, fmt.Errorf("scan movie genre: %w", err)
		}

		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate movie genres: %w", err)
	}

	return result, nil
}

func (r *MovieRepo) getMovieActors(ctx context.Context, movieID int64) ([]domain.ActorShort, error) {
	rows, err := r.db.Query(ctx, sqlGetMovieActorsByID, movieID)
	if err != nil {
		return nil, fmt.Errorf("get movie actors: %w", err)
	}
	defer rows.Close()

	result := make([]domain.ActorShort, 0)

	for rows.Next() {
		var actor domain.ActorShort
		if err = rows.Scan(&actor.ID, &actor.FullName, &actor.PictureFileKey); err != nil {
			return nil, fmt.Errorf("scan movie actor: %w", err)
		}

		result = append(result, actor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate movie actors: %w", err)
	}

	return result, nil
}

func (r *MovieRepo) getMovieEpisodes(ctx context.Context, movieID int64) ([]domain.Episode, error) {
	rows, err := r.db.Query(ctx, sqlGetMovieEpisodesByID, movieID)
	if err != nil {
		return nil, fmt.Errorf("get movie episodes: %w", err)
	}
	defer rows.Close()

	result := make([]domain.Episode, 0)

	for rows.Next() {
		var episode domain.Episode
		if err = rows.Scan(
			&episode.ID,
			&episode.MovieID,
			&episode.SeasonNumber,
			&episode.EpisodeNumber,
			&episode.Title,
			&episode.Description,
			&episode.DurationSeconds,
			&episode.PictureFileKey,
			&episode.VideoFileKey,
		); err != nil {
			return nil, fmt.Errorf("scan movie episode: %w", err)
		}

		result = append(result, episode)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate movie episodes: %w", err)
	}

	return result, nil
}

func (r *MovieRepo) getGenreMovies(ctx context.Context, genreID int64) ([]domain.MovieCard, error) {
	rows, err := r.db.Query(ctx, sqlGetGenreMoviesByID, genreID)
	if err != nil {
		return nil, fmt.Errorf("get genre movies: %w", err)
	}
	defer rows.Close()

	result := make([]domain.MovieCard, 0)

	for rows.Next() {
		var movie domain.MovieCard
		if err = rows.Scan(&movie.ID, &movie.Title, &movie.PictureFileKey); err != nil {
			return nil, fmt.Errorf("scan genre movie: %w", err)
		}

		result = append(result, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate genre movies: %w", err)
	}

	return result, nil
}

func (r *MovieRepo) getActorMovies(ctx context.Context, actorID int64) ([]domain.MovieCard, error) {
	rows, err := r.db.Query(ctx, sqlGetActorMoviesByID, actorID)
	if err != nil {
		return nil, fmt.Errorf("get actor movies: %w", err)
	}
	defer rows.Close()

	result := make([]domain.MovieCard, 0)

	for rows.Next() {
		var movie domain.MovieCard
		if err = rows.Scan(&movie.ID, &movie.Title, &movie.PictureFileKey); err != nil {
			return nil, fmt.Errorf("scan actor movie: %w", err)
		}

		result = append(result, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate actor movies: %w", err)
	}

	return result, nil
}
