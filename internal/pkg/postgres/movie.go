package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/jackc/pgx/v5"
)

type MovieRepo struct {
	db *Client
}

func NewMovieRepo(db *Client) *MovieRepo {
	return &MovieRepo{db: db}
}

var (
	ErrSelectionNotFound = errors.New("selection not found")
	ErrMovieNotFound     = errors.New("movie not found")
	ErrActorNotFound     = errors.New("actor not found")
	ErrEpisodeNotFound   = errors.New("episode not found")
)

func (r *MovieRepo) GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error) {
	rows, err := r.db.Pool.Query(ctx, sqlGetSelectionByTitle, title)
	if err != nil {
		return domain.SelectionResponse{}, fmt.Errorf("unable to query selections: %w", err)
	}
	defer rows.Close()

	var moviePreviews []domain.MoviePreview

	for rows.Next() {
		var moviePreview domain.MoviePreview

		err := rows.Scan(&moviePreview.ID, &moviePreview.Title, &moviePreview.ImgUrl)
		if err != nil {
			return domain.SelectionResponse{}, fmt.Errorf("unable to read moviePreview: %w", err)
		}

		moviePreviews = append(moviePreviews, moviePreview)
	}

	return domain.SelectionResponse{
		Title:  title,
		Movies: moviePreviews,
	}, nil
}

func (r *MovieRepo) GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error) {
	rows, err := r.db.Pool.Query(ctx, sqlGetAllSelectionTitles)
	if err != nil {
		return nil, fmt.Errorf("unable to query selection titles: %w", err)
	}
	defer rows.Close()

	var selectionTitles []string

	for rows.Next() {
		var selectionTitle string

		err := rows.Scan(&selectionTitle)
		if err != nil {
			return nil, fmt.Errorf("unable to read selection title: %w", err)
		}

		selectionTitles = append(selectionTitles, selectionTitle)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating selection titles: %w", err)
	}

	var selections []domain.SelectionResponse

	for _, title := range selectionTitles {
		selection, err := r.GetSelectionByTitle(ctx, title)
		if err != nil {
			return nil, fmt.Errorf("unable to get selection by title %s: %w", title, err)
		}

		selections = append(selections, selection)
	}

	return selections, nil
}

func (r *MovieRepo) GetMovieByID(ctx context.Context, id int64) (domain.MovieResponse, error) {
	var movieResponse domain.MovieResponse

	err := r.db.Pool.QueryRow(ctx, sqlGetMovieByID, id).Scan(
		&movieResponse.ID,
		&movieResponse.Title,
		&movieResponse.Description,
		&movieResponse.Director,
		&movieResponse.ContentType,
		&movieResponse.ReleaseYear,
		&movieResponse.DurationSeconds,
		&movieResponse.AgeLimit,
		&movieResponse.OriginalLanguageID,
		&movieResponse.CountryID,
		&movieResponse.PictureFileKey,
		&movieResponse.PosterFileKey,
	)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	genres, err := r.getGenresByMovieID(ctx, id)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	actors, err := r.getActorsByMovieID(ctx, id)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	movieResponse.Genres = genres
	movieResponse.Actors = actors

	return movieResponse, nil
}

func (r *MovieRepo) getGenresByMovieID(ctx context.Context, movieID int64) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, sqlGetGenresByMovieID, movieID)
	if err != nil {
		return nil, fmt.Errorf("unable to query genres by movie id: %w", err)
	}
	defer rows.Close()

	genres := make([]string, 0)
	for rows.Next() {
		var genre string

		if err = rows.Scan(&genre); err != nil {
			return nil, fmt.Errorf("unable to scan genre: %w", err)
		}

		genres = append(genres, genre)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating genres: %w", err)
	}

	return genres, nil
}

func (r *MovieRepo) getActorsByMovieID(ctx context.Context, movieID int64) ([]domain.ActorPreview, error) {
	rows, err := r.db.Pool.Query(ctx, sqlGetActorsByMovieID, movieID)
	if err != nil {
		return nil, fmt.Errorf("unable to query actors by movie id: %w", err)
	}
	defer rows.Close()

	actors := make([]domain.ActorPreview, 0)
	for rows.Next() {
		var actor domain.ActorPreview

		if err = rows.Scan(&actor.ID, &actor.FullName, &actor.PictureFileKey); err != nil {
			return nil, fmt.Errorf("unable to scan actor preview: %w", err)
		}

		actors = append(actors, actor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating actors: %w", err)
	}

	return actors, nil
}

func (r *MovieRepo) GetActorByID(ctx context.Context, id int64) (domain.ActorResponse, error) {
	var actor domain.ActorResponse

	err := r.db.Pool.QueryRow(ctx, sqlGetActorByID, id).Scan(
		&actor.ID,
		&actor.FullName,
		&actor.BirthDate,
		&actor.Biography,
		&actor.CountryID,
		&actor.PictureFileKey,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ActorResponse{}, ErrActorNotFound
		}

		return domain.ActorResponse{}, fmt.Errorf("unable to scan actor: %w", err)
	}

	movies, err := r.getMoviesByActorID(ctx, id)
	if err != nil {
		return domain.ActorResponse{}, err
	}

	actor.Movies = movies

	return actor, nil
}

func (r *MovieRepo) getMoviesByActorID(ctx context.Context, actorID int64) ([]domain.MoviePreview, error) {
	rows, err := r.db.Pool.Query(ctx, sqlGetMoviesByActorID, actorID)
	if err != nil {
		return nil, fmt.Errorf("unable to query movies by actor id: %w", err)
	}
	defer rows.Close()

	movies := make([]domain.MoviePreview, 0)
	for rows.Next() {
		var movie domain.MoviePreview

		if err = rows.Scan(&movie.ID, &movie.Title, &movie.ImgUrl); err != nil {
			return nil, fmt.Errorf("unable to scan movie preview: %w", err)
		}

		movies = append(movies, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating movies: %w", err)
	}

	return movies, nil
}

// GetEpisodesByMovieID Получаем все эпизоды связанные с фильмом
func (r *MovieRepo) GetEpisodesByMovieID(ctx context.Context, movieID int64) ([]domain.EpisodeItemResponse, error) {
	rows, err := r.db.Pool.Query(ctx, sqlGetEpisodesByMovieID, movieID)
	if err != nil {
		return nil, fmt.Errorf("unable to query episodes by movie id: %w", err)
	}
	defer rows.Close()

	var episodes []domain.EpisodeItemResponse

	for rows.Next() {
		var episode domain.EpisodeItemResponse

		err = rows.Scan(
			&episode.ID,
			&episode.MovieID,
			&episode.SeasonNumber,
			&episode.EpisodeNumber,
			&episode.Title,
			&episode.Description,
			&episode.DurationSeconds,
			&episode.ImgURL,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan episode item: %w", err)
		}

		episodes = append(episodes, episode)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating episodes: %w", err)
	}

	return episodes, nil
}

// GetEpisodePlayback Получаем данные о запускаемом эпизоде
func (r *MovieRepo) GetEpisodePlayback(ctx context.Context, episodeID int64) (domain.EpisodePlaybackResponse, error) {
	var playback domain.EpisodePlaybackResponse

	err := r.db.Pool.QueryRow(ctx, sqlGetEpisodePlayback, episodeID).Scan(
		&playback.EpisodeID,
		&playback.MovieID,
		&playback.SeasonNumber,
		&playback.EpisodeNumber,
		&playback.Title,
		&playback.DurationSeconds,
		&playback.PlaybackURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.EpisodePlaybackResponse{}, ErrEpisodeNotFound
		}

		return domain.EpisodePlaybackResponse{}, fmt.Errorf("unable to scan episode playback: %w", err)
	}

	return playback, nil
}

func (r *MovieRepo) GetWatchProgress(ctx context.Context, userID, episodeID int64) (int, error) {
	var positionSeconds int

	err := r.db.Pool.QueryRow(ctx, sqlGetWatchProgress, userID, episodeID).Scan(&positionSeconds)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		return 0, fmt.Errorf("unable to get watch progress: %w", err)
	}

	return positionSeconds, nil
}

func (r *MovieRepo) UpsertWatchProgress(ctx context.Context, userID, episodeID int64, positionSeconds int) error {
	_, err := r.db.Pool.Exec(ctx, sqlUpsertWatchProgress, userID, episodeID, positionSeconds)
	if err != nil {
		return fmt.Errorf("unable to upsert watch progress: %w", err)
	}

	return nil
}
