package domain

import (
	"time"

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

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// TODO сделать на дженериках

// func (u *User) Name() string {
//	return "users"
// }
//
// func UserSerialize(user User) ([]byte, error) {
//	return serializer.Serialize(user)
// }
//
// func UserDeserialize(data []byte, user *User) error {
//	return serializer.Deserialize(data, user)
// }
