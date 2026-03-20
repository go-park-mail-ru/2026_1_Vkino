package domain

import (
	"errors"
)

// в usecase только использую ошибку на уровне репозитория.
var (
	ErrBadSelectionTitle = errors.New("selection with this title doesn't exist")
)
