package domain

import (
	"time"

	validator "github.com/go-park-mail-ru/2026_1_VKino/pkg/validatex"
)

type User struct {
	ID               int64
	Email            string
	Password         string
	Birthdate        *time.Time
	AvatarFileKey    *string
	RegistrationDate time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func ValidateEmailQuery(query string) bool {
	return validator.ValidateEmailQuery(query)
}
