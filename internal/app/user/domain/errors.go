package domain

import (
	"errors"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoSession          = errors.New("no session")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidBirthdate   = errors.New("invalid birthdate")
	ErrInvalidAvatar      = errors.New("invalid avatar")
	ErrPasswordMismatch   = errors.New("password mismatch")
	ErrInternal           = errors.New("this error now not exists")
)
