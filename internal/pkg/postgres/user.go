package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	db *Postgres
}

func NewUserRepo(db *Postgres) *UserRepo {
	return &UserRepo{db: db}
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	sql := `select id, email, password_hash, registration_date, is_active, created_at, updated_at 
	from users where email = $1`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, sql, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.RegistrationDate,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		// TODO выяснить какую ошибку вернет постгрес при already exists
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	sql := `select id, email, password_hash, registration_date, is_active, created_at, updated_at 
	from users where id = $1`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, sql, id).Scan(
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
	now := time.Now()
	userID := uuid.New()

	sql := `insert into users (id, email, password_hash, registration_date, is_active, created_at, updated_at) 
	values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Pool.Exec(ctx, sql,
		userID,
		email,
		passwordHash,
		now,
		true,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &domain.User{
		ID:               userID,
		Email:            email,
		Password:         passwordHash,
		RegistrationDate: now,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	sql := `
		UPDATE users 
		SET password_hash = $1, updated_at = $2 
		WHERE email = $3 
		RETURNING id, email, password_hash, registration_date, is_active, created_at, updated_at
	`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, sql, passwordHash, time.Now(), email).Scan(
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
	sql := `delete from users where email = $1`

	tag, err := r.db.Pool.Exec(ctx, sql, email)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
