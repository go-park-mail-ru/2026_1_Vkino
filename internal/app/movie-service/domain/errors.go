package domain

import "errors"

var (
	ErrMovieNotFound         = errors.New("movie not found")
	ErrActorNotFound         = errors.New("actor not found")
	ErrSelectionNotFound     = errors.New("selection not found")
	ErrGenreNotFound         = errors.New("genre not found")
	ErrEpisodeNotFound       = errors.New("episode not found")
	ErrInvalidMovieID        = errors.New("invalid movie id")
	ErrInvalidActorID        = errors.New("invalid actor id")
	ErrInvalidEpisodeID      = errors.New("invalid episode id")
	ErrInvalidSelectionTitle = errors.New("invalid selection title")
	ErrInvalidGenreID        = errors.New("invalid genre id")
	ErrInvalidSearchQuery    = errors.New("invalid search query")
	ErrInvalidWatchProgress  = errors.New("invalid watch progress")
	ErrInternal              = errors.New("internal error")
)
