package db

import (
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/serializer"
	"github.com/google/uuid"
)

func init() {
	serializer.RegisterType(User{})
	serializer.RegisterType(time.Time{})
	serializer.RegisterType(uuid.UUID{})
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this login already exists")
)

type User struct {
	ID               uuid.UUID
	Login            string
	Password         string
	RegistrationDate time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (u *User) Name() string {
	return "users"
}

func UserSerialize(value User) ([]byte, error) {
	return serializer.Serialize(value)
}

func UserDeserialize(data []byte, value *User) error {
	return serializer.Deserialize(data, value)
}

func (repo *Repository) GetUserByLogin(login string) (*User, error) {
	repo.mu.RLock()
	userData, exists := repo.tbls["users"][login]
	repo.mu.RUnlock()

	if !exists {
		return nil, ErrUserNotFound
	}

	var user User
	if err := UserDeserialize(userData, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *Repository) GetUserByID(id uuid.UUID) (*User, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	users := repo.tbls["users"]

	for _, userData := range users {
		var user User
		if err := UserDeserialize(userData, &user); err != nil {
			return nil, err
		}

		if user.ID == id {
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (repo *Repository) CreateUser(login string, password string) (*User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.tbls["users"][login]; exists {
		return nil, ErrUserAlreadyExists
	}

	now := time.Now()

	user := User{
		ID:               uuid.New(),
		Login:            login,
		Password:         password,
		RegistrationDate: now,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	data, err := UserSerialize(user)
	if err != nil {
		return nil, err
	}

	repo.tbls["users"][login] = data

	return &user, nil
}

func (repo *Repository) UpdateUser(login string, password string) (*User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	userData, exists := repo.tbls["users"][login]
	if !exists {
		return nil, ErrUserNotFound
	}

	var user User
	err := UserDeserialize(userData, &user)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	updatedUser := User{
		ID:               user.ID,
		Login:            login,
		Password:         password,
		RegistrationDate: user.RegistrationDate,
		IsActive:         user.IsActive,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        now,
	}

	data, err := UserSerialize(updatedUser)
	if err != nil {
		return nil, err
	}

	repo.tbls["users"][login] = data
	return &updatedUser, nil
}

func (repo *Repository) GetUsers() ([]*User, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	var users []*User

	for _, userData := range repo.tbls["users"] {
		var user User
		if err := UserDeserialize(userData, &user); err != nil {
			return nil, err
		}

		if user.IsActive {
			users = append(users, &user)
		}
	}

	return users, nil
}

func (repo *Repository) DeleteUser(login string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	userData, exists := repo.tbls["users"][login]
	if !exists {
		return ErrUserNotFound
	}

	var user User
	if err := UserDeserialize(userData, &user); err != nil {
		return err
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	data, err := UserSerialize(user)
	if err != nil {
		return err
	}

	repo.tbls["users"][login] = data
	return nil
}
