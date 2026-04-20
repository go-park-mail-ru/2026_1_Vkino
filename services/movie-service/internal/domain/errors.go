package domain

import "errors"

var (
	ErrMovieNotFound         = errors.New("movie not found")
	ErrActorNotFound         = errors.New("actor not found")
	ErrSelectionNotFound     = errors.New("selection not found")
	ErrInvalidMovieID        = errors.New("invalid movie id")
	ErrInvalidActorID        = errors.New("invalid actor id")
	ErrInvalidSelectionTitle = errors.New("invalid selection title")
	ErrInternal              = errors.New("internal server error")
)
