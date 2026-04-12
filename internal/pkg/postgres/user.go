package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
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

	err := r.db.Pool.QueryRow(ctx, sqlGetUserByEmail, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
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

	err := r.db.Pool.QueryRow(ctx, sqlGetUserByID, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
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

	err := r.db.Pool.QueryRow(ctx, sqlCreateUser, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
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

	err := r.db.Pool.QueryRow(ctx, sqlUpdateUser, passwordHash, time.Now(), email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
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
	tag, err := r.db.Pool.Exec(ctx, sqlDeleteUser, email)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
