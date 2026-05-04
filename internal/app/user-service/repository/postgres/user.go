//nolint:gocyclo,lll // Repository methods are kept explicit and close to their SQL contracts.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
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
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrMovieNotFound     = errors.New("movie not found")
)

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

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
			return nil, domain.ErrUserNotFound
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
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) SearchUsersByEmail(
	ctx context.Context,
	userID int64,
	query string,
) ([]domain.UserSearchResult, error) {
	rows, err := r.db.Query(ctx, sqlSearchUsersByEmail, userID, query)
	if err != nil {
		return nil, fmt.Errorf("search users by email: %w", err)
	}
	defer rows.Close()

	users := make([]domain.UserSearchResult, 0)

	for rows.Next() {
		var user domain.UserSearchResult

		if err = rows.Scan(&user.ID, &user.Email, &user.AvatarURL, &user.IsFriend); err != nil {
			return nil, fmt.Errorf("scan searched users: %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate searched users: %w", err)
	}

	return users, nil
}

func (r *UserRepo) UpdateBirthdate(ctx context.Context, userID int64, birthdate *time.Time) (*domain.User, error) {
	var user domain.User

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
			return nil, domain.ErrUserNotFound
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
			return nil, domain.ErrUserNotFound
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

func (r *UserRepo) GetFavorites(ctx context.Context, userID int64, limit, offset int32) ([]int64, int32, error) {
	rows, err := r.db.Query(ctx, sqlGetFavorites, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get favorites: %w", err)
	}
	defer rows.Close()

	movieIDs := make([]int64, 0, limit)

	for rows.Next() {
		var movieID int64
		if err := rows.Scan(&movieID); err != nil {
			return nil, 0, fmt.Errorf("scan favorite movie id: %w", err)
		}

		movieIDs = append(movieIDs, movieID)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate favorites: %w", err)
	}

	var total int32
	if err := r.db.QueryRow(ctx, sqlCountFavorites, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}

	return movieIDs, total, nil
}

func (r *UserRepo) AddFriend(ctx context.Context, userID int64, friendID int64) error {
	u1, u2 := orderedFriendPair(userID, friendID)

	_, err := r.db.Exec(ctx, sqlAddFriend, u1, u2)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyFriends
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
		return domain.ErrFriendNotFound
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
		return 0, domain.ErrAlreadyFriends
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

//nolint:funcorder // Transaction helper is intentionally kept next to the calling flow.
func (r *UserRepo) acceptFriendRequestTx(ctx context.Context, requestID, toUserID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin accept friend request tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	var fromUserID, rowToUserID int64

	err = tx.QueryRow(ctx, sqlAcceptFriendRequestUpdate, requestID, toUserID).Scan(&fromUserID, &rowToUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrFriendNotFound
		}

		return fmt.Errorf("accept friend request update: %w", err)
	}

	u1, u2 := orderedFriendPair(fromUserID, rowToUserID)
	if _, err := tx.Exec(ctx, sqlAcceptFriendInsert, u1, u2); err != nil {
		return fmt.Errorf("accept friend request insert friend: %w", err)
	}

	if _, err := tx.Exec(ctx, sqlAcceptFriendDeleteRequest, requestID); err != nil {
		return fmt.Errorf("accept friend request delete row: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("accept friend request commit: %w", err)
	}

	return nil
}

func (r *UserRepo) RespondToFriendRequest(ctx context.Context, requestID, userID int64, action string) error {
	if action == "accept" {
		return r.acceptFriendRequestTx(ctx, requestID, userID)
	}

	if action == "cancel" {
		return r.DeleteOutgoingFriendRequest(ctx, requestID, userID)
	}

	var fromUserID int64

	err := r.db.QueryRow(ctx, sqlRespondToRequest, "declined", requestID, userID).Scan(&fromUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrFriendNotFound
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
			return domain.ErrFriendNotFound
		}

		return fmt.Errorf("delete outgoing friend request: %w", err)
	}

	return nil
}

func (r *UserRepo) GetFriendRequests(ctx context.Context, userID int64, direction string, limit int32) ([]domain.FriendRequestItem, error) {
	query := sqlGetIncomingRequests
	if direction == "outgoing" {
		query = sqlGetOutgoingRequests
	}

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get friend requests: %w", err)
	}
	defer rows.Close()

	items := make([]domain.FriendRequestItem, 0, limit)

	for rows.Next() {
		var item domain.FriendRequestItem

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

func (r *UserRepo) GetFriendsList(ctx context.Context, userID int64, limit, offset int32) ([]domain.UserSearchResult, int32, error) {
	rows, err := r.db.Query(ctx, sqlGetFriends, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get friends list: %w", err)
	}
	defer rows.Close()

	friends := make([]domain.UserSearchResult, 0, limit)

	for rows.Next() {
		var friend domain.UserSearchResult
		if err := rows.Scan(&friend.ID, &friend.Email, &friend.AvatarURL); err != nil {
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
			return "", domain.ErrUserNotFound
		}

		return "", fmt.Errorf("get user role: %w", err)
	}

	return role, nil
}

func ignoreRollbackError(err error) {
	if err != nil {
		return
	}
}
