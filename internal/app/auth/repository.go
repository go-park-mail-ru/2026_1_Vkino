package auth

import (
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/db"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/serializer"

	"github.com/google/uuid"
)

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

func UserSerialize(user User) ([]byte, error) {
	return serializer.Serialize(user)
}

func UserDeserialize(data []byte, user *User) error {
	return serializer.Deserialize(data, user)
}

type Repository struct {
	db *db.DB
}

func NewRepository(db *db.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetUserByLogin(login string) (*User, error) {
	data, err := r.db.Get("users", login)
	if err != nil {
		if err == db.ErrNotFound || err == db.ErrTableNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var user User
	if err := UserDeserialize(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByID(id uuid.UUID) (*User, error) {
	allData, err := r.db.GetAll("users")
	if err != nil {
		if err == db.ErrTableNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	for _, data := range allData {
		var user User
		if err := UserDeserialize(data, &user); err != nil {
			return nil, err
		}

		if user.ID == id {
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *Repository) CreateUser(login string, password string) (*User, error) {
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

	err = r.db.Save("users", login, data)
	if err != nil {
		if err == db.ErrAlreadyExists {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) UpdateUser(login string, password string) (*User, error) {
	user, err := r.GetUserByLogin(login)
	if err != nil {
		return nil, err
	}

	user.Password = password
	user.UpdatedAt = time.Now()

	data, err := UserSerialize(*user)
	if err != nil {
		return nil, err
	}

	if err := r.db.Update("users", login, data); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetUsers() ([]*User, error) {
	allData, err := r.db.GetAll("users")
	if err != nil {
		if err == db.ErrTableNotFound {
			return []*User{}, nil
		}
		return nil, err
	}

	var users []*User
	for _, data := range allData {
		var user User
		if err := UserDeserialize(data, &user); err != nil {
			return nil, err
		}

		if user.IsActive {
			users = append(users, &user)
		}
	}

	return users, nil
}

func (r *Repository) DeleteUser(login string) error {
	err := r.db.Delete("users", login)
	if err != nil {
		if err == db.ErrNotFound || err == db.ErrTableNotFound {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}
