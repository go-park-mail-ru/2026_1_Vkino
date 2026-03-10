package auth

import "errors"

var (
	ErrReadingConfig     = errors.New("error reading config file")
	ErrUmarshalingConfig = errors.New("error marshaling config")
)
