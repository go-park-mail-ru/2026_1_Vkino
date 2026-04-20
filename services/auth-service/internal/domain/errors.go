package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoSession          = errors.New("no session")
	ErrInvalidToken       = errors.New("invalid token")
	ErrPasswordMismatch   = errors.New("password mismatch")
	ErrInternal           = errors.New("this error now not exists")
)