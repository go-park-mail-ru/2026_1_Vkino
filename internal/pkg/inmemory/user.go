package inmemory

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	profiledomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
)

type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this login already exists")
)

func (r *UserRepo) GetUserByEmail(email string) (*domain.User, error) {
	data, err := r.db.Get("users", email)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrTableNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	var user domain.User
	if err := serializer.Deserialize(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetUserByID(id int64) (*domain.User, error) {
	allData, err := r.db.GetAll("users")
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	for _, data := range allData {
		var user domain.User
		if err := serializer.Deserialize(data, &user); err != nil {
			return nil, err
		}

		if user.ID == id {
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}

func (r *UserRepo) GetProfileByID(_ context.Context, id int64) (profiledomain.ProfileResponse, error) {
	user, err := r.GetUserByID(id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return profiledomain.ProfileResponse{}, profiledomain.ErrUserNotFound
		}

		return profiledomain.ProfileResponse{}, err
	}

	return profiledomain.ProfileResponse{
		Email: user.Email,
	}, nil
}

func (r *UserRepo) CreateUser(login string, password string) (*domain.User, error) {
	now := time.Now()

	user := domain.User{
		ID:               now.UnixNano(),
		Email:            login,
		Password:         password,
		RegistrationDate: now,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	data, err := serializer.Serialize(user)
	if err != nil {
		return nil, err
	}

	err = r.db.Save("users", login, data)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) UpdateUser(login string, password string) (*domain.User, error) {
	user, err := r.GetUserByEmail(login)
	if err != nil {
		return nil, err
	}

	user.Password = password
	user.UpdatedAt = time.Now()

	data, err := serializer.Serialize(*user)
	if err != nil {
		return nil, err
	}

	if err := r.db.Update("users", login, data); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepo) GetAllUsers() ([]*domain.User, error) {
	allData, err := r.db.GetAll("users")
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return []*domain.User{}, nil
		}

		return nil, err
	}

	var users []*domain.User

	for _, data := range allData {
		var user domain.User
		if err := serializer.Deserialize(data, &user); err != nil {
			return nil, err
		}

		if user.IsActive {
			users = append(users, &user)
		}
	}

	return users, nil
}

func (r *UserRepo) DeleteUser(login string) error {
	err := r.db.Delete("users", login)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrTableNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	return nil
}
