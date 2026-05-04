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

func Validate(email, password string) bool {
	return validator.ValidateEmail(email) && validator.ValidatePassword(password)
}

func ValidatePassword(password string) bool {
	return validator.ValidatePassword(password)
}
