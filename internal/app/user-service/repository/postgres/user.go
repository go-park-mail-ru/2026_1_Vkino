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
		&user.Role,
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
		&user.Role,
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
		&user.Role,
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
		&user.Role,
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

func (r *UserRepo) ToggleFavorite(ctx context.Context, userID, movieID int64) (bool, error) {
	var isFavorite bool
	err := r.db.QueryRow(ctx, sqlToggleFavorite, userID, movieID).Scan(&isFavorite)
	if err != nil {
		return false, fmt.Errorf("toggle favorite: %w", err)
	}
	return isFavorite, nil
}

func (r *UserRepo) GetFavorites(ctx context.Context, userID int64, limit, offset int32) ([]domain2.MovieCardResponse, int32, error) {
	rows, err := r.db.Query(ctx, sqlGetFavorites, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get favorites: %w", err)
	}
	defer rows.Close()

	movies := make([]domain2.MovieCardResponse, 0, limit)
	for rows.Next() {
		var movie domain2.MovieCardResponse
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.PictureFileKey); err != nil {
			return nil, 0, fmt.Errorf("scan favorite movie: %w", err)
		}
		movies = append(movies, movie)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate favorites: %w", err)
	}

	var total int32
	if err := r.db.QueryRow(ctx, sqlCountFavorites, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}

	return movies, total, nil
}

func (r *UserRepo) AddFriend(ctx context.Context, userID int64, friendID int64) error {
	u1, u2 := orderedFriendPair(userID, friendID)
	_, err := r.db.Exec(ctx, sqlAddFriend, u1, u2)
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
	u1, u2 := orderedFriendPair(userID, friendID)
	tag, err := r.db.Exec(ctx, sqlDeleteFriend, u1, u2)
	if err != nil {
		return fmt.Errorf("delete friend: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain2.ErrFriendNotFound
	}

	if _, err := r.db.Exec(ctx, sqlDeleteFriendRequestsBetweenUsers, userID, friendID); err != nil {
		return fmt.Errorf("cleanup friend requests after delete friend: %w", err)
	}

	return nil
}

func (r *UserRepo) SendFriendRequest(ctx context.Context, fromUserID, toUserID int64) (int64, error) {
	p1, p2 := orderedFriendPair(fromUserID, toUserID)
	var areFriends bool
	if err := r.db.QueryRow(ctx, sqlAreFriends, p1, p2).Scan(&areFriends); err != nil {
		return 0, fmt.Errorf("check friends before request: %w", err)
	}
	if areFriends {
		return 0, domain2.ErrAlreadyFriends
	}

	var status string
	err := r.db.QueryRow(ctx, sqlGetFriendRequestStatus, fromUserID, toUserID).Scan(&status)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("get friend request status: %w", err)
	}
	if err == nil && status == "accepted" {
		if _, delErr := r.db.Exec(ctx, sqlDeleteFriendRequestPair, fromUserID, toUserID); delErr != nil {
			return 0, fmt.Errorf("cleanup accepted friend request: %w", delErr)
		}
	}

	var requestID int64
	err = r.db.QueryRow(ctx, sqlSendFriendRequest, fromUserID, toUserID).Scan(&requestID)
	if err != nil {
		return 0, fmt.Errorf("send friend request: %w", err)
	}
	return requestID, nil
}

func (r *UserRepo) RespondToFriendRequest(ctx context.Context, requestID, userID int64, action string) error {
	if action == "accept" {
		var fromUserID int64
		err := r.db.QueryRow(ctx, sqlAcceptFriendRequestAtomic, requestID, userID).Scan(&fromUserID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain2.ErrFriendNotFound
			}
			return fmt.Errorf("accept friend request atomically: %w", err)
		}
		return nil
	}

	if action == "cancel" {
		return r.DeleteOutgoingFriendRequest(ctx, requestID, userID)
	}

	var fromUserID int64
	err := r.db.QueryRow(ctx, sqlRespondToRequest, "declined", requestID, userID).Scan(&fromUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain2.ErrFriendNotFound
		}
		return fmt.Errorf("decline friend request: %w", err)
	}

	return nil
}

func (r *UserRepo) DeleteOutgoingFriendRequest(ctx context.Context, requestID, fromUserID int64) error {
	var toUserID int64
	err := r.db.QueryRow(ctx, sqlDeleteOutgoingRequest, requestID, fromUserID).Scan(&toUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain2.ErrFriendNotFound
		}
		return fmt.Errorf("delete outgoing friend request: %w", err)
	}
	return nil
}

func (r *UserRepo) GetFriendRequests(ctx context.Context, userID int64, direction string, limit int32) ([]domain2.FriendRequestItem, error) {
	query := sqlGetIncomingRequests
	if direction == "outgoing" {
		query = sqlGetOutgoingRequests
	}

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get friend requests: %w", err)
	}
	defer rows.Close()

	items := make([]domain2.FriendRequestItem, 0, limit)
	for rows.Next() {
		var item domain2.FriendRequestItem
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &item.Email, &createdAt); err != nil {
			return nil, fmt.Errorf("scan friend request: %w", err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate friend requests: %w", err)
	}
	return items, nil
}

func (r *UserRepo) GetFriendsList(ctx context.Context, userID int64, limit, offset int32) ([]domain2.UserSearchResult, int32, error) {
	rows, err := r.db.Query(ctx, sqlGetFriends, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get friends list: %w", err)
	}
	defer rows.Close()

	friends := make([]domain2.UserSearchResult, 0, limit)
	for rows.Next() {
		var friend domain2.UserSearchResult
		if err := rows.Scan(&friend.ID, &friend.Email); err != nil {
			return nil, 0, fmt.Errorf("scan friend: %w", err)
		}
		friend.IsFriend = true
		friends = append(friends, friend)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate friends: %w", err)
	}

	var total int32
	if err := r.db.QueryRow(ctx, sqlCountFriends, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count friends: %w", err)
	}

	return friends, total, nil
}

func (r *UserRepo) GetUserRole(ctx context.Context, userID int64) (string, error) {
	var role string

	err := r.db.QueryRow(ctx, sqlGetUserRole, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}

		return "", fmt.Errorf("get user role: %w", err)
	}

	return role, nil
}
