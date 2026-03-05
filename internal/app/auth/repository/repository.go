package repository

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/google/uuid"
)

type UserRepo interface {
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByID(id uuid.UUID) (*domain.User, error)
	CreateUser(login string, password string) (*domain.User, error)
	UpdateUser(login string, password string) (*domain.User, error)
	GetAllUsers() ([]*domain.User, error)
	DeleteUser(login string) error
}

type SessionRepo interface {
	SaveSession(email string, tokens domain.TokenPair) error
	GetSession(email string) (*domain.TokenPair, error)
	DeleteSession(email string) error
}
