package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		&user.CredentialHash,
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
		&user.CredentialHash,
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

func (r *UserRepo) GetActiveSubscription(
	ctx context.Context,
	userID int64,
) (domain.SubscriptionInfo, error) {
	var (
		info      domain.SubscriptionInfo
		expiresAt time.Time
	)

	err := r.db.QueryRow(ctx, sqlGetActiveSubscription, userID).Scan(
		&info.ID,
		&info.Code,
		&info.Name,
		&info.Level,
		&expiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SubscriptionInfo{}, nil
		}

		return domain.SubscriptionInfo{}, fmt.Errorf("get active subscription: %w", err)
	}

	activeUntil := expiresAt.Format(time.RFC3339)
	info.ActiveUntil = &activeUntil

	return info, nil
}

func (r *UserRepo) GetSubscriptionTariffByCode(
	ctx context.Context,
	code string,
) (domain.SubscriptionInfo, error) {
	var info domain.SubscriptionInfo

	err := r.db.QueryRow(ctx, sqlGetSubscriptionTariffByCode, code).Scan(
		&info.ID,
		&info.Code,
		&info.Name,
		&info.Level,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SubscriptionInfo{}, nil
		}

		return domain.SubscriptionInfo{}, fmt.Errorf("get subscription tariff by code: %w", err)
	}

	return info, nil
}

func (r *UserRepo) GetSubscriptionTariffByID(
	ctx context.Context,
	tariffID int64,
) (domain.SubscriptionTariff, error) {
	var tariff domain.SubscriptionTariff

	err := r.db.QueryRow(ctx, sqlGetSubscriptionTariffByID, tariffID).Scan(
		&tariff.ID,
		&tariff.Code,
		&tariff.Title,
		&tariff.Level,
		&tariff.DurationDays,
		&tariff.PriceVKinoCoins,
		&tariff.IsCoinsPaymentAvailable,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SubscriptionTariff{}, domain.ErrSubscriptionTariffNotFound
		}

		return domain.SubscriptionTariff{}, fmt.Errorf("get subscription tariff by id: %w", err)
	}

	return tariff, nil
}

func (r *UserRepo) DeactivateUserSubscriptions(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, sqlDeactivateUserSubscriptions, userID)
	if err != nil {
		return fmt.Errorf("deactivate user subscriptions: %w", err)
	}

	return nil
}

func (r *UserRepo) CreateUserSubscription(
	ctx context.Context,
	userID, tariffID int64,
	startsAt, expiresAt time.Time,
) error {
	_, err := r.db.Exec(ctx, sqlCreateUserSubscription, userID, tariffID, startsAt, expiresAt)
	if err != nil {
		return fmt.Errorf("create user subscription: %w", err)
	}

	return nil
}

func (r *UserRepo) GetSubscriptionTariffOptions(
	ctx context.Context,
	tariffID int64,
) ([]domain.SubscriptionOption, error) {
	rows, err := r.db.Query(ctx, sqlGetSubscriptionTariffOptions, tariffID)
	if err != nil {
		return nil, fmt.Errorf("get subscription tariff options: %w", err)
	}
	defer rows.Close()

	options := make([]domain.SubscriptionOption, 0)

	for rows.Next() {
		var (
			option domain.SubscriptionOption
			value  sql.NullString
		)

		if err = rows.Scan(&option.Code, &value); err != nil {
			return nil, fmt.Errorf("scan subscription tariff option: %w", err)
		}

		if value.Valid {
			optionValue := value.String
			option.Value = &optionValue
		}

		options = append(options, option)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscription tariff options: %w", err)
	}

	return options, nil
}

func (r *UserRepo) GetCoinsReceivedToday(ctx context.Context, userID int64) (int32, error) {
	var total int32
	if err := r.db.QueryRow(ctx, sqlGetCoinsReceivedToday, userID).Scan(&total); err != nil {
		return 0, fmt.Errorf("get coins received today: %w", err)
	}

	return total, nil
}

func (r *UserRepo) GetVKinoCoinsBalance(ctx context.Context, userID int64) (int32, error) {
	var balance int32
	if err := r.db.QueryRow(ctx, sqlGetVKinoCoinsBalance, userID).Scan(&balance); err != nil {
		return 0, fmt.Errorf("get vkino coins balance: %w", err)
	}

	return balance, nil
}

func (r *UserRepo) GrantDailyVKinoCoins(ctx context.Context, userID int64) (int32, error) {
	var granted int32
	if err := r.db.QueryRow(ctx, sqlGrantDailyVKinoCoins, userID).Scan(&granted); err != nil {
		return 0, fmt.Errorf("grant daily vkino coins: %w", err)
	}

	return granted, nil
}

func (r *UserRepo) GetRoomsCreatedThisMonth(ctx context.Context, userID int64) (int32, error) {
	var total int32
	if err := r.db.QueryRow(ctx, sqlGetRoomsCreatedThisMonth, userID).Scan(&total); err != nil {
		return 0, fmt.Errorf("get rooms created this month: %w", err)
	}

	return total, nil
}

func (r *UserRepo) GetVKinoCoinsHistory(
	ctx context.Context,
	userID int64,
	limit, offset int32,
) ([]domain.VKinoCoinsHistoryItem, int32, error) {
	rows, err := r.db.Query(ctx, sqlGetVKinoCoinsHistory, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get vkino coins history: %w", err)
	}
	defer rows.Close()

	items := make([]domain.VKinoCoinsHistoryItem, 0, limit)

	var totalCount int32

	for rows.Next() {
		var (
			id              sql.NullInt64
			vkinoCoinsCount sql.NullInt32
			operationType   sql.NullString
			description     sql.NullString
			createdAt       sql.NullTime
			rowTotalCount   int32
		)

		if err := rows.Scan(
			&id,
			&vkinoCoinsCount,
			&operationType,
			&description,
			&createdAt,
			&rowTotalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan vkino coins history item: %w", err)
		}

		totalCount = rowTotalCount

		if !id.Valid {
			continue
		}

		item := domain.VKinoCoinsHistoryItem{
			ID:              id.Int64,
			VKinoCoinsCount: vkinoCoinsCount.Int32,
			OperationType:   operationType.String,
			Description:     description.String,
		}

		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format(time.RFC3339)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate vkino coins history: %w", err)
	}

	return items, totalCount, nil
}

func (r *UserRepo) SpendVKinoCoins(
	ctx context.Context,
	userID int64,
	coinsAmount int32,
	operationType string,
	description string,
) (int32, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin spend vkino coins tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	if err = lockUserForCoinsPurchase(ctx, tx, userID); err != nil {
		return 0, err
	}

	balance, err := getSufficientVKinoCoinsBalanceTx(ctx, tx, userID, coinsAmount)
	if err != nil {
		return 0, err
	}

	if _, err = tx.Exec(ctx, sqlCreateVKinoCoinsHistory, userID, coinsAmount, operationType, description); err != nil {
		return 0, fmt.Errorf("create vkino coins history: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit spend vkino coins tx: %w", err)
	}

	return balance - coinsAmount, nil
}

func (r *UserRepo) GrantVKinoCoins(
	ctx context.Context,
	userID int64,
	coinsAmount int32,
	operationType string,
	description string,
	referenceKey string,
) (int32, int32, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("begin grant vkino coins tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	if err = lockUserForCoinsPurchase(ctx, tx, userID); err != nil {
		return 0, 0, err
	}

	tag, err := tx.Exec(
		ctx,
		sqlCreateVKinoCoinsGrantHistory,
		userID,
		coinsAmount,
		operationType,
		description,
		referenceKey,
	)
	if err != nil {
		return 0, 0, fmt.Errorf("create vkino coins grant history: %w", err)
	}

	balance, err := getVKinoCoinsBalanceTx(ctx, tx, userID)
	if err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("commit grant vkino coins tx: %w", err)
	}

	coinsGranted := int32(0)
	if tag.RowsAffected() > 0 {
		coinsGranted = coinsAmount
	}

	return coinsGranted, balance, nil
}

func (r *UserRepo) BuySubscriptionWithVKinoCoins(
	ctx context.Context,
	userID int64,
	tariffID int64,
) (domain.VKinoCoinsSubscriptionPurchase, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.VKinoCoinsSubscriptionPurchase{}, fmt.Errorf("begin buy subscription with vkino coins tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	tariff, balance, now, err := prepareVKinoCoinsSubscriptionPurchase(ctx, tx, userID, tariffID)
	if err != nil {
		return domain.VKinoCoinsSubscriptionPurchase{}, err
	}

	paymentID, err := createCoinsPaymentTx(ctx, tx, userID, tariff.ID, tariff.PriceVKinoCoins, now)
	if err != nil {
		return domain.VKinoCoinsSubscriptionPurchase{}, err
	}

	if err = createVKinoCoinsPurchaseHistoryTx(ctx, tx, userID, tariff); err != nil {
		return domain.VKinoCoinsSubscriptionPurchase{}, err
	}

	activeUntil, err := activatePurchasedSubscriptionTx(ctx, tx, userID, tariff, now)
	if err != nil {
		return domain.VKinoCoinsSubscriptionPurchase{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.VKinoCoinsSubscriptionPurchase{}, fmt.Errorf("commit buy subscription with vkino coins tx: %w", err)
	}

	return domain.VKinoCoinsSubscriptionPurchase{
		PaymentID:         paymentID,
		CoinsSpent:        tariff.PriceVKinoCoins,
		VKinoCoinsBalance: balance - tariff.PriceVKinoCoins,
		Subscription: domain.SubscriptionInfo{
			ID:          tariff.ID,
			Code:        tariff.Code,
			Name:        tariff.Title,
			Level:       tariff.Level,
			ActiveUntil: activeUntil,
		},
	}, nil
}

func prepareVKinoCoinsSubscriptionPurchase(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	tariffID int64,
) (domain.SubscriptionTariff, int32, time.Time, error) {
	tariff, err := getAvailableCoinsTariffByIDTx(ctx, tx, tariffID)
	if err != nil {
		return domain.SubscriptionTariff{}, 0, time.Time{}, err
	}

	if err = lockUserForCoinsPurchase(ctx, tx, userID); err != nil {
		return domain.SubscriptionTariff{}, 0, time.Time{}, err
	}

	balance, err := getSufficientVKinoCoinsBalanceTx(ctx, tx, userID, tariff.PriceVKinoCoins)
	if err != nil {
		return domain.SubscriptionTariff{}, 0, time.Time{}, err
	}

	return tariff, balance, time.Now().UTC(), nil
}

func getAvailableCoinsTariffByIDTx(
	ctx context.Context,
	tx pgx.Tx,
	tariffID int64,
) (domain.SubscriptionTariff, error) {
	tariff, err := getSubscriptionTariffByIDTx(ctx, tx, tariffID)
	if err != nil {
		return domain.SubscriptionTariff{}, err
	}

	if !tariff.IsCoinsPaymentAvailable || tariff.PriceVKinoCoins <= 0 {
		return domain.SubscriptionTariff{}, domain.ErrTariffNotAvailableForCoins
	}

	return tariff, nil
}

func getSubscriptionTariffByIDTx(
	ctx context.Context,
	tx pgx.Tx,
	tariffID int64,
) (domain.SubscriptionTariff, error) {
	var tariff domain.SubscriptionTariff

	err := tx.QueryRow(ctx, sqlGetSubscriptionTariffByID, tariffID).Scan(
		&tariff.ID,
		&tariff.Code,
		&tariff.Title,
		&tariff.Level,
		&tariff.DurationDays,
		&tariff.PriceVKinoCoins,
		&tariff.IsCoinsPaymentAvailable,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SubscriptionTariff{}, domain.ErrSubscriptionTariffNotFound
		}

		return domain.SubscriptionTariff{}, fmt.Errorf("get subscription tariff by id in tx: %w", err)
	}

	return tariff, nil
}

func lockUserForCoinsPurchase(ctx context.Context, tx pgx.Tx, userID int64) error {
	var lockedUserID int64
	if err := tx.QueryRow(ctx, sqlLockUserForUpdate, userID).Scan(&lockedUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}

		return fmt.Errorf("lock user for coins purchase: %w", err)
	}

	return nil
}

func getVKinoCoinsBalanceTx(ctx context.Context, tx pgx.Tx, userID int64) (int32, error) {
	var balance int32
	if err := tx.QueryRow(ctx, sqlGetVKinoCoinsBalance, userID).Scan(&balance); err != nil {
		return 0, fmt.Errorf("get vkino coins balance in tx: %w", err)
	}

	return balance, nil
}

func getSufficientVKinoCoinsBalanceTx(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	requiredBalance int32,
) (int32, error) {
	balance, err := getVKinoCoinsBalanceTx(ctx, tx, userID)
	if err != nil {
		return 0, err
	}

	if balance < requiredBalance {
		return 0, domain.ErrInsufficientVKinoCoins
	}

	return balance, nil
}

func createCoinsPaymentTx(
	ctx context.Context,
	tx pgx.Tx,
	userID, tariffID int64,
	coinsSpent int32,
	paidAt time.Time,
) (int64, error) {
	amount := fmt.Sprintf("%d.00", coinsSpent)

	var (
		paymentID int64
		createdAt time.Time
		updatedAt time.Time
	)

	err := tx.QueryRow(
		ctx,
		sqlCreateCoinsPayment,
		userID,
		"subscription",
		tariffID,
		amount,
		"VKC",
		"succeeded",
		uuid.NewString(),
		"vkino_coins",
		coinsSpent,
		paidAt,
	).Scan(&paymentID, &createdAt, &updatedAt)
	if err != nil {
		return 0, fmt.Errorf("create coins payment: %w", err)
	}

	return paymentID, nil
}

func createVKinoCoinsPurchaseHistoryTx(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	tariff domain.SubscriptionTariff,
) error {
	description := "Покупка подписки " + tariff.Code

	if _, err := tx.Exec(
		ctx,
		sqlCreateVKinoCoinsPurchaseHistory,
		userID,
		tariff.PriceVKinoCoins,
		description,
	); err != nil {
		return fmt.Errorf("create vkino coins purchase history: %w", err)
	}

	return nil
}

func activatePurchasedSubscriptionTx(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	tariff domain.SubscriptionTariff,
	now time.Time,
) (*string, error) {
	startsAt, err := subscriptionActivationStartTx(ctx, tx, userID, tariff, now)
	if err != nil {
		return nil, err
	}

	expiresAt := startsAt.AddDate(0, 0, int(tariff.DurationDays))

	if _, err = tx.Exec(ctx, sqlDeactivateUserSubscriptions, userID); err != nil {
		return nil, fmt.Errorf("deactivate user subscriptions: %w", err)
	}

	if _, err = tx.Exec(ctx, sqlCreateUserSubscription, userID, tariff.ID, startsAt, expiresAt); err != nil {
		return nil, fmt.Errorf("create user subscription: %w", err)
	}

	activeUntil := expiresAt.Format(time.RFC3339)

	return &activeUntil, nil
}

func subscriptionActivationStartTx(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	tariff domain.SubscriptionTariff,
	now time.Time,
) (time.Time, error) {
	var (
		current   domain.SubscriptionInfo
		expiresAt time.Time
	)

	err := tx.QueryRow(ctx, sqlGetActiveSubscription, userID).Scan(
		&current.ID,
		&current.Code,
		&current.Name,
		&current.Level,
		&expiresAt,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, fmt.Errorf("get active subscription in tx: %w", err)
	}

	startsAt := now
	if err == nil && current.Level == tariff.Level && expiresAt.After(now) {
		startsAt = expiresAt
	}

	return startsAt, nil
}

func (r *UserRepo) GetFriend(ctx context.Context, userID, friendID int64) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx, sqlGetFriendByID, userID, friendID).Scan(
		&user.ID,
		&user.Email,
		&user.CredentialHash,
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
			return nil, domain.ErrFriendNotFound
		}

		return nil, fmt.Errorf("get friend by id: %w", err)
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
		&user.CredentialHash,
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
		&user.CredentialHash,
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

func (r *UserRepo) SetMovieRating(ctx context.Context, userID, movieID int64, rating float64) error {
	tag, err := r.db.Exec(ctx, sqlUpsertUserMovieRating, userID, movieID, rating)
	if err != nil {
		return fmt.Errorf("upsert user movie rating: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrMovieNotFound
	}

	return nil
}

func (r *UserRepo) SetMovieReview(
	ctx context.Context,
	userID, movieID int64,
	rating *float64,
	comment *string,
) (domain.MovieReviewResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.MovieReviewResponse{}, fmt.Errorf("begin set movie review tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	var (
		reviewResp domain.MovieReviewResponse
		dbRating   sql.NullFloat64
		dbComment  sql.NullString
	)

	err = queryMovieReview(ctx, tx, userID, movieID, rating, comment, &reviewResp, &dbRating, &dbComment)
	if err != nil {
		return domain.MovieReviewResponse{}, handleSetMovieReviewError(err)
	}

	if err = deleteReviewReactionsIfCommentMissing(ctx, tx, dbComment, reviewResp.ReviewID); err != nil {
		return domain.MovieReviewResponse{}, fmt.Errorf("cleanup review reactions: %w", err)
	}

	assignMovieReviewResponseFields(&reviewResp, dbRating, dbComment)

	if err = tx.Commit(ctx); err != nil {
		return domain.MovieReviewResponse{}, fmt.Errorf("commit set movie review tx: %w", err)
	}

	return reviewResp, nil
}

func handleSetMovieReviewError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrMovieNotFound
	}

	return fmt.Errorf("upsert movie review: %w", err)
}

func deleteReviewReactionsIfCommentMissing(
	ctx context.Context,
	tx pgx.Tx,
	comment sql.NullString,
	reviewID int64,
) error {
	if comment.Valid && comment.String != "" {
		return nil
	}

	_, err := tx.Exec(ctx, sqlDeleteReviewReactionsByReviewID, reviewID)

	return err
}

func assignMovieReviewResponseFields(
	reviewResp *domain.MovieReviewResponse,
	rating sql.NullFloat64,
	comment sql.NullString,
) {
	if rating.Valid {
		ratingValue := rating.Float64
		reviewResp.Rating = &ratingValue
	}

	if comment.Valid && comment.String != "" {
		commentValue := comment.String
		reviewResp.Comment = &commentValue
	}
}

func queryMovieReview(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	movieID int64,
	rating *float64,
	comment *string,
	reviewResp *domain.MovieReviewResponse,
	dbRating *sql.NullFloat64,
	dbComment *sql.NullString,
) error {
	return tx.QueryRow(ctx, sqlUpsertUserMovieReview, userID, movieID, rating, comment).Scan(
		&reviewResp.ReviewID,
		&reviewResp.MovieID,
		dbRating,
		dbComment,
	)
}

func (r *UserRepo) DeleteMovieReview(ctx context.Context, userID, movieID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete movie review tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	var (
		reviewID   int64
		isFavorite bool
	)

	err = tx.QueryRow(ctx, sqlGetReviewByUserAndMovie, userID, movieID).Scan(&reviewID, &isFavorite)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("get movie review by user and movie: %w", err)
	}

	if _, err = tx.Exec(ctx, sqlDeleteReviewReactionsByReviewID, reviewID); err != nil {
		return fmt.Errorf("delete movie review reactions: %w", err)
	}

	if err = removeMovieReviewRow(ctx, tx, reviewID, isFavorite); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete movie review tx: %w", err)
	}

	return nil
}

func removeMovieReviewRow(ctx context.Context, tx pgx.Tx, reviewID int64, isFavorite bool) error {
	if isFavorite {
		if _, err := tx.Exec(ctx, sqlClearUserMovieReview, reviewID); err != nil {
			return fmt.Errorf("clear movie review: %w", err)
		}

		return nil
	}

	if _, err := tx.Exec(ctx, sqlDeleteUserMovieReviewRow, reviewID); err != nil {
		return fmt.Errorf("delete movie review row: %w", err)
	}

	return nil
}

func (r *UserRepo) SetReviewReaction(ctx context.Context, userID, reviewID int64, reaction string) error {
	tag, err := r.db.Exec(ctx, sqlSetReviewReaction, userID, reviewID, reaction)
	if err != nil {
		return fmt.Errorf("set review reaction: %w", err)
	}

	if tag.RowsAffected() > 0 {
		return nil
	}

	var (
		reviewOwnerID int64
		hasComment    bool
	)

	err = r.db.QueryRow(ctx, sqlGetReviewOwner, reviewID).Scan(&reviewOwnerID, &hasComment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrMovieReviewNotFound
		}

		return fmt.Errorf("get review owner: %w", err)
	}

	if reviewOwnerID == userID {
		return domain.ErrSelfReviewVote
	}

	if !hasComment {
		return domain.ErrMovieReviewNotFound
	}

	return domain.ErrInternal
}

func (r *UserRepo) DeleteReviewReaction(ctx context.Context, userID, reviewID int64) error {
	if _, err := r.db.Exec(ctx, sqlDeleteReviewReaction, reviewID, userID); err != nil {
		return fmt.Errorf("delete review reaction: %w", err)
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

	var total int32

	for rows.Next() {
		var (
			movieID sql.NullInt64
			count   int32
		)
		if err := rows.Scan(&movieID, &count); err != nil {
			return nil, 0, fmt.Errorf("scan favorite movie id: %w", err)
		}

		total = count

		if movieID.Valid {
			movieIDs = append(movieIDs, movieID.Int64)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate favorites: %w", err)
	}

	return movieIDs, total, nil
}

func (r *UserRepo) AddFriend(ctx context.Context, userID int64, friendID int64) (*domain.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin add friend tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	if _, err = getUserByIDTx(ctx, tx, userID); err != nil {
		return nil, err
	}

	friend, err := prepareFriendForAddTx(ctx, tx, friendID)
	if err != nil {
		return nil, err
	}

	if err = createFriendshipTx(ctx, tx, userID, friendID); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit add friend tx: %w", err)
	}

	return friend, nil
}

func (r *UserRepo) DeleteFriend(ctx context.Context, userID int64, friendID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete friend tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	u1, u2 := orderedFriendPair(userID, friendID)

	tag, err := tx.Exec(ctx, sqlDeleteFriend, u1, u2)
	if err != nil {
		return fmt.Errorf("delete friend: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrFriendNotFound
	}

	if _, err = tx.Exec(ctx, sqlDeleteFriendRequestsBetweenUsers, userID, friendID); err != nil {
		return fmt.Errorf("cleanup friend requests after delete friend: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete friend tx: %w", err)
	}

	return nil
}

func (r *UserRepo) SendFriendRequest(ctx context.Context, fromUserID, toUserID int64) (int64, error) {
	p1, p2 := orderedFriendPair(fromUserID, toUserID)

	areFriends, err := r.areFriends(ctx, p1, p2)
	if err != nil {
		return 0, err
	}

	if areFriends {
		return 0, domain.ErrAlreadyFriends
	}

	if err = r.cleanupAcceptedFriendRequest(ctx, fromUserID, toUserID); err != nil {
		return 0, err
	}

	var requestID int64

	err = r.db.QueryRow(ctx, sqlSendFriendRequest, fromUserID, toUserID).Scan(&requestID)
	if err != nil {
		if isForeignKeyViolation(
			err,
			friendRequestFromUserIDForeignKey,
			friendRequestToUserIDForeignKey,
		) {
			return 0, domain.ErrUserNotFound
		}

		return 0, fmt.Errorf("send friend request: %w", err)
	}

	return requestID, nil
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

func (r *UserRepo) GetFriendRequests(
	ctx context.Context,
	userID int64,
	direction string,
	limit int32,
) ([]domain.FriendRequestItem, error) {
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

func (r *UserRepo) GetFriendsList(
	ctx context.Context,
	userID int64,
	limit, offset int32,
) ([]domain.UserSearchResult, int32, error) {
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

func (r *UserRepo) areFriends(ctx context.Context, user1ID, user2ID int64) (bool, error) {
	var areFriends bool
	if err := r.db.QueryRow(ctx, sqlAreFriends, user1ID, user2ID).Scan(&areFriends); err != nil {
		return false, fmt.Errorf("check friends before request: %w", err)
	}

	return areFriends, nil
}

func (r *UserRepo) cleanupAcceptedFriendRequest(ctx context.Context, fromUserID, toUserID int64) error {
	status, exists, err := r.friendRequestStatus(ctx, fromUserID, toUserID)
	if err != nil {
		return err
	}

	if !exists || status != "accepted" {
		return nil
	}

	if _, err = r.db.Exec(ctx, sqlDeleteFriendRequestPair, fromUserID, toUserID); err != nil {
		return fmt.Errorf("cleanup accepted friend request: %w", err)
	}

	return nil
}

func (r *UserRepo) friendRequestStatus(ctx context.Context, fromUserID, toUserID int64) (string, bool, error) {
	var status string

	err := r.db.QueryRow(ctx, sqlGetFriendRequestStatus, fromUserID, toUserID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("get friend request status: %w", err)
	}

	return status, true, nil
}

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

func ignoreRollbackError(err error) {
	if err != nil {
		return
	}
}

func prepareFriendForAddTx(ctx context.Context, tx pgx.Tx, friendID int64) (*domain.User, error) {
	friend, err := getUserByIDTx(ctx, tx, friendID)
	if err != nil {
		return nil, err
	}

	return friend, nil
}

func createFriendshipTx(ctx context.Context, tx pgx.Tx, userID int64, friendID int64) error {
	u1, u2 := orderedFriendPair(userID, friendID)

	if _, err := tx.Exec(ctx, sqlAddFriend, u1, u2); err != nil {
		return mapAddFriendError(err)
	}

	if _, err := tx.Exec(ctx, sqlDeleteFriendRequestsBetweenUsers, userID, friendID); err != nil {
		return fmt.Errorf("cleanup friend requests after add friend: %w", err)
	}

	return nil
}

func mapAddFriendError(err error) error {
	switch {
	case isUniqueConstraintViolation(err, friendUniqueConstraint):
		return domain.ErrAlreadyFriends
	case isForeignKeyViolation(err, friendUser1IDForeignKey, friendUser2IDForeignKey):
		return domain.ErrUserNotFound
	default:
		return fmt.Errorf("add friend: %w", err)
	}
}

func getUserByIDTx(ctx context.Context, tx pgx.Tx, userID int64) (*domain.User, error) {
	var user domain.User

	err := tx.QueryRow(ctx, sqlGetUserByID, userID).Scan(
		&user.ID,
		&user.Email,
		&user.CredentialHash,
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

		return nil, fmt.Errorf("get user by id in tx: %w", err)
	}

	return &user, nil
}
