package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepo struct {
	db *corepostgres.Client
}

func NewUserRepo(db *corepostgres.Client) *UserRepo {
	return &UserRepo{db: db}
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrMovieNotFound     = errors.New("movie not found")
)

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain2.User, error) {
	var user domain2.User

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

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*domain2.User, error) {
	var user domain2.User

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

func (r *UserRepo) SearchUsersByEmail(ctx context.Context, userID int64,
	query string) ([]domain2.UserSearchResult, error) {
	rows, err := r.db.Query(ctx, sqlSearchUsersByEmail, userID, query)
	if err != nil {
		return nil, fmt.Errorf("search users by email: %w", err)
	}
	defer rows.Close()

	users := make([]domain2.UserSearchResult, 0)

	for rows.Next() {
		var user domain2.UserSearchResult

		if err = rows.Scan(&user.ID, &user.Email, &user.IsFriend); err != nil {
			return nil, fmt.Errorf("scan searched users: %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate searched users: %w", err)
	}

	return users, nil
}

func (r *UserRepo) UpdateBirthdate(ctx context.Context, userID int64, birthdate *time.Time) (*domain2.User, error) {
	var user domain2.User

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

func (r *UserRepo) UpdateAvatarFileKey(ctx context.Context, userID int64, avatarFileKey *string) (*domain2.User, error) {
	var user domain2.User

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

func (r *UserRepo) AddFriend(ctx context.Context, userID int64, friendID int64) error {
	_, err := r.db.Exec(ctx, sqlAddFriend, userID, friendID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain2.ErrAlreadyFriends
		}

		return fmt.Errorf("add friend: %w", err)
	}

	return nil
}

func (r *UserRepo) DeleteFriend(ctx context.Context, userID int64, friendID int64) error {
	tag, err := r.db.Exec(ctx, sqlDeleteFriend, userID, friendID)
	if err != nil {
		return fmt.Errorf("delete friend: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain2.ErrFriendNotFound
	}

	return nil
}
