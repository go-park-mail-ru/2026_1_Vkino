package usecase

import "errors"

var (
	ErrBcryptGenerate = errors.New("bcrypt generate error")
	ErrTokenGenerate  = errors.New("token generate error")
	ErrSessionSave    = errors.New("failed to save session")
)
