package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MovieRepo struct {
	db *Postgres
}

func NewMovieRepo(db *Postgres) *MovieRepo {
	return &MovieRepo{db: db}
}

var (
	ErrSelectionNotFound = errors.New("selection not found")
	ErrMovieNotFound     = errors.New("movie not found")
	ErrActorNotFound     = errors.New("actor not found")
)

func (r *MovieRepo) GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error) {
	sql := `
select m.id, m.title, m.picture_file_key
from movie m 
join movie_to_selection mts ON (mts.movie_id = m.id)
where mts.selection_id = (select id from selection where title=$1)
`
	// нужен ли таймаут внутренний?

	rows, err := r.db.Pool.Query(ctx, sql, title)
	fmt.Println(err)
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

	var selection domain.SelectionResponse
	selection.Title = title
	selection.Movies = moviePreviews

	return selection, nil
}

func (r *MovieRepo) GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error) {
	sqlTitles := "select title from selection"

	fmt.Println("start selections")

	rows, err := r.db.Pool.Query(ctx, sqlTitles)

	fmt.Println("query ok?", err)

	if err != nil {
		return nil, fmt.Errorf("unable to query selection titles: %w", err)
	}
	defer rows.Close()

	var selectionTitles []string

	for rows.Next() {
		var selectionTitle string
		err := rows.Scan(&selectionTitle)
		fmt.Println("scanning rows", err)
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
		fmt.Println(err)
		if err != nil {
			return nil, fmt.Errorf("unable to get selection by title %s: %w", title, err)
		}
		selections = append(selections, selection)
	}
	fmt.Println(selections)
	return selections, nil
}

func (r *MovieRepo) GetMovieByID(ctx context.Context, id uuid.UUID) (domain.MovieResponse, error) {
	sql := `select title, description, director, content_type, release_year, 
    duration_seconds, age_limit, original_language_id, country_id, picture_file_key from movie where id=$1`

	var movieResponse domain.MovieResponse
	err := r.db.Pool.QueryRow(ctx, sql, id).Scan(
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

func (r *MovieRepo) GetActorByID(ctx context.Context, id uuid.UUID) (domain.ActorResponse, error) {
	sql := `
	SELECT 
		id,
		full_name, 
		birthdate, 
		biography, 
		country_id, 
		picture_file_key 
	FROM actor 
	WHERE id = $1
	`

	var actor domain.ActorResponse
	err := r.db.Pool.QueryRow(ctx, sql, id).Scan(
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
