package main

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this login already exists")
)

type User struct {
	ID               int32
	Login            string
	Password         string
	RegistrationDate time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Repository struct {
	mu   sync.RWMutex
	tbls map[string]map[string][]byte
}

func InitRepo() *Repository {
	return &Repository{
		tbls: map[string]map[string][]byte{
			"users": make(map[string][]byte),
		},
	}
}

func deserialize(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

func serialize(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (repo *Repository) GetUserByLogin(login string) (*User, error) {
	repo.mu.RLock()
	userData, exists := repo.tbls["users"][login]
	repo.mu.RUnlock()

	if !exists {
		return nil, ErrUserNotFound
	}

	var user User
	if err := deserialize(userData, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *Repository) CreateUser(login string, password string) (*User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.tbls["users"][login]; exists {
		return nil, ErrUserAlreadyExists
	}

	now := time.Now()

	user := &User{
		ID:               int32(len(repo.tbls["users"]) + 1),
		Login:            login,
		Password:         password,
		RegistrationDate: now,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	data, err := serialize(user)
	if err != nil {
		return nil, err
	}

	repo.tbls["users"][login] = data

	return user, nil
}

func (repo *Repository) UpdateUser(login string, password string) (*User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	userData, exists := repo.tbls["users"][login]
	if !exists {
		return nil, ErrUserNotFound
	}

	var user User
	err := deserialize(userData, &user)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	updatedUser := &User{
		ID:               user.ID,
		Login:            login,
		Password:         password,
		RegistrationDate: user.RegistrationDate,
		IsActive:         user.IsActive,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        now,
	}

	data, err := serialize(updatedUser)
	if err != nil {
		return nil, err
	}

	repo.tbls["users"][login] = data
	return updatedUser, nil
}

func (repo *Repository) GetUsers() ([]*User, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	var users []*User

	for _, userData := range repo.tbls["users"] {
		var user User
		if err := deserialize(userData, &user); err != nil {
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
	if err := deserialize(userData, &user); err != nil {
		return err
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	data, err := serialize(user)
	if err != nil {
		return err
	}

	repo.tbls["users"][login] = data
	return nil
}

func (repo *Repository) FillMockData() {
	repo.CreateUser("user1", "123")
	repo.CreateUser("user2", "234")
	repo.CreateUser("user3", "456")
	repo.CreateUser("user4", "456")
	repo.CreateUser("user5", "456")

}
