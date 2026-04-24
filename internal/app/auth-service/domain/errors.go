package domain

import (
	"errors"

	jwtsvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/jwt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoSession          = errors.New("no session")
	ErrInvalidToken       = jwtsvc.ErrInvalidToken
	ErrPasswordMismatch   = errors.New("password mismatch")
	ErrInternal           = errors.New("internal error")
)
