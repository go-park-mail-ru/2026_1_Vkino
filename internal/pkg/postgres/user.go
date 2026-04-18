package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepo struct {
	db *Client
}

func NewUserRepo(db *Client) *UserRepo {
	return &UserRepo{db: db}
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx, sqlGetUserByEmail, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Birthdate,
		&user.AvatarFileKey,
		&user.RegistrationDate,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx, sqlGetUserByID, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Birthdate,
		&user.AvatarFileKey,
		&user.RegistrationDate,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx, sqlCreateUser, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Birthdate,
		&user.AvatarFileKey,
		&user.RegistrationDate,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx, sqlUpdateUser, passwordHash, time.Now(), email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Birthdate,
		&user.AvatarFileKey,
		&user.RegistrationDate,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, email string) error {
	tag, err := r.db.Exec(ctx, sqlDeleteUser, email)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepo) UpdateBirthdate(ctx context.Context, userID int64, birthdate *time.Time) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, sqlUpdateUserBirthdate, birthdate, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Birthdate,
		&user.AvatarFileKey,
		&user.RegistrationDate,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user birthdate: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) UpdateAvatarFileKey(ctx context.Context, userID int64, avatarFileKey *string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, sqlUpdateUserAvatarFileKey, avatarFileKey, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Birthdate,
		&user.AvatarFileKey,
		&user.RegistrationDate,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user avatar key: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	tag, err := r.db.Exec(ctx, sqlUpdateUserPasswordByID, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update user password by id: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepo) AddMovieToFavorites(ctx context.Context, userID, movieID int64) error {
	tag, err := r.db.Exec(ctx, sqlUpsertUserFavoriteMovie, userID, movieID)
	if err != nil {
		return fmt.Errorf("upsert user favorite movie: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrMovieNotFound
	}

	return nil
}
