package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/domain"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	db *corepostgres.Client
}

func NewUserRepo(db *corepostgres.Client) *UserRepo {
	return &UserRepo{db: db}
}

const (
	signupBonusCoinsAmount = 20
	signupBonusDescription = "Стартовый бонус за регистрацию"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx, sqlGetUserByEmail, email).Scan(
		&user.ID,
		&user.Email,
		&user.CredentialHash,
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
		&user.CredentialHash,
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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create user transaction: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	var user domain.User

	err = tx.QueryRow(ctx, sqlCreateUser, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.CredentialHash,
		&user.Birthdate,
		&user.AvatarFileKey,
		&user.RegistrationDate,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if isUniqueConstraintViolation(err, userEmailUniqueConstraint) {
			return nil, ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	if _, err = tx.Exec(
		ctx,
		sqlCreateSignupBonusHistory,
		user.ID,
		signupBonusCoinsAmount,
		signupBonusDescription,
	); err != nil {
		return nil, fmt.Errorf("create signup bonus history: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create user transaction: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	tag, err := r.db.Exec(ctx, sqlUpdateUserHashByID, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update user password by id: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func ignoreRollbackError(err error) {
	if err != nil {
		return
	}
}
