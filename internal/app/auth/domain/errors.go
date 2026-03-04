package domain

import (
	stderrors "errors"
)

var (
	ErrUserAlreadyExists  = stderrors.New("user already exists")
	ErrInvalidCredentials = stderrors.New("invalid credentials")
	ErrNoSession          = stderrors.New("no session")
	ErrInvalidToken       = stderrors.New("invalid token")
	ErrOther              = stderrors.New("this error now not exists")
)
