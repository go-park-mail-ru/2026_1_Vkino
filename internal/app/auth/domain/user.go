package domain

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID
	Email            string
	Password         string
	RegistrationDate time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (u *User) Name() string {
	return "users"
}

func Serialize[T any](v T) ([]byte, error) {
	return serializer.Serialize(v)
}

func Deserialize[T any](data []byte, v *T) error {
	return serializer.Deserialize(data, v)
}
