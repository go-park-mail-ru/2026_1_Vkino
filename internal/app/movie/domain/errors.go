package domain

import (
	"errors"
)

// в usecase только использую ошибку на уровне репозитория.
var (
	ErrBadSelectionTitle    = errors.New("selection with this title doesn't exist")
	ErrInvalidMovieID       = errors.New("invalid movie id")
	ErrInvalidActorID       = errors.New("invalid actor id")
	ErrInvalidEpisodeID     = errors.New("invalid episode id")
	ErrInvalidWatchProgress = errors.New("invalid watch progress")
)
