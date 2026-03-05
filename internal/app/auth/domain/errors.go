package domain

import (
	"errors"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoSession          = errors.New("no session")
	ErrInvalidToken       = errors.New("invalid token")
	ErrOther              = errors.New("this error now not exists")
)
