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
		err := rows.Scan(&moviePreview.ID, &moviePreview.Title, &moviePreview.PictureFileKey)
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
	)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	return movieResponse, nil
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

	return actor, nil
}
