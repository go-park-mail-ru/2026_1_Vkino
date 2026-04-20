package domain

import (
	"errors"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSearchQuery = errors.New("invalid search query")
	ErrNoSession          = errors.New("no session")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidMovieID     = errors.New("invalid movie id")
	ErrInvalidBirthdate   = errors.New("invalid birthdate")
	ErrInvalidAvatar      = errors.New("invalid avatar")
	ErrPasswordMismatch   = errors.New("password mismatch")
	ErrAlreadyFriends     = errors.New("already friends")
	ErrFriendNotFound     = errors.New("friend not found")
	ErrSelfFriendship     = errors.New("self friendship is forbidden")
	ErrInternal           = errors.New("this error now not exists")
)
