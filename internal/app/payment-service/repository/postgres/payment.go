package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/jackc/pgx/v5"
)

type PaymentRepo struct {
	db *corepostgres.Client
}

func NewPaymentRepo(db *corepostgres.Client) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) UserExists(ctx context.Context, userID int64) (bool, error) {
	var exists bool

	if err := r.db.QueryRow(ctx, sqlUserExists, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}

	return exists, nil
}

func (r *PaymentRepo) GetSubscriptionTariff(
	ctx context.Context,
	tariffID int64,
) (domain.SubscriptionTariff, error) {
	var tariff domain.SubscriptionTariff

	err := r.db.QueryRow(ctx, sqlGetSubscriptionTariff, tariffID).Scan(
		&tariff.ID,
		&tariff.Code,
		&tariff.Title,
		&tariff.PriceMoney,
		&tariff.PriceVKinoCoins,
		&tariff.IsCoinsPaymentAvailable,
		&tariff.IsMoneyPaymentAvailable,
		&tariff.DurationDays,
		&tariff.Level,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SubscriptionTariff{}, domain.ErrTariffNotFound
		}

		return domain.SubscriptionTariff{}, fmt.Errorf("get subscription tariff: %w", err)
	}

	return tariff, nil
}

func (r *PaymentRepo) GetCoinsPack(ctx context.Context, packID int64) (domain.CoinsPack, error) {
	var pack domain.CoinsPack

	err := r.db.QueryRow(ctx, sqlGetCoinsPack, packID).Scan(
		&pack.ID,
		&pack.Code,
		&pack.Title,
		&pack.CoinsAmount,
		&pack.PriceMoney,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CoinsPack{}, domain.ErrCoinsPackNotFound
		}

		return domain.CoinsPack{}, fmt.Errorf("get coins pack: %w", err)
	}

	return pack, nil
}

func (r *PaymentRepo) ListCoinsPacks(ctx context.Context) ([]domain.CoinsPack, error) {
	rows, err := r.db.Query(ctx, sqlListCoinsPacks)
	if err != nil {
		return nil, fmt.Errorf("list coins packs: %w", err)
	}
	defer rows.Close()

	packs := make([]domain.CoinsPack, 0)

	for rows.Next() {
		var pack domain.CoinsPack

		if err = rows.Scan(
			&pack.ID,
			&pack.Code,
			&pack.Title,
			&pack.CoinsAmount,
			&pack.PriceMoney,
		); err != nil {
			return nil, fmt.Errorf("scan coins pack: %w", err)
		}

		packs = append(packs, pack)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate coins packs: %w", err)
	}

	return packs, nil
}

func (r *PaymentRepo) ListMoneyTariffs(ctx context.Context) ([]domain.MoneyTariff, error) {
	rows, err := r.db.Query(ctx, sqlListMoneyTariffs)
	if err != nil {
		return nil, fmt.Errorf("list money tariffs: %w", err)
	}
	defer rows.Close()

	tariffs := make([]domain.MoneyTariff, 0)

	for rows.Next() {
		var tariff domain.MoneyTariff

		if err = rows.Scan(
			&tariff.ID,
			&tariff.Code,
			&tariff.Title,
			&tariff.PriceMoney,
			&tariff.PriceVKinoCoins,
			&tariff.IsCoinsPaymentAvailable,
			&tariff.IsMoneyPaymentAvailable,
			&tariff.DurationDays,
			&tariff.Level,
		); err != nil {
			return nil, fmt.Errorf("scan money tariff: %w", err)
		}

		tariffs = append(tariffs, tariff)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate money tariffs: %w", err)
	}

	return tariffs, nil
}

func (r *PaymentRepo) CreatePayment(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	var coinsSpent any
	if payment.CoinsSpent != nil {
		coinsSpent = *payment.CoinsSpent
	}

	err := r.db.QueryRow(
		ctx,
		sqlCreatePayment,
		payment.UserID,
		string(payment.ProductType),
		payment.ProductRefID,
		payment.Amount,
		payment.Currency,
		string(payment.Status),
		payment.IdempotencyKey,
		string(payment.PaymentMethod),
		coinsSpent,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("create payment: %w", err)
	}

	return payment, nil
}

func (r *PaymentRepo) UpdatePaymentYooKassa(
	ctx context.Context,
	paymentID int64,
	yookassaPaymentID, confirmationURL string,
) error {
	_, err := r.db.Exec(ctx, sqlUpdatePaymentYooKassa, paymentID, yookassaPaymentID, confirmationURL)
	if err != nil {
		return fmt.Errorf("update payment yookassa: %w", err)
	}

	return nil
}

func (r *PaymentRepo) UpdatePaymentStatus(
	ctx context.Context,
	paymentID int64,
	status domain.PaymentStatus,
	paidAt *time.Time,
) error {
	_, err := r.db.Exec(ctx, sqlUpdatePaymentStatus, paymentID, string(status), paidAt)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	return nil
}

func (r *PaymentRepo) GetPaymentByID(ctx context.Context, paymentID int64) (domain.Payment, error) {
	return r.scanPayment(r.db.QueryRow(ctx, sqlGetPaymentByID, paymentID))
}

func (r *PaymentRepo) GetPaymentByYooKassaID(
	ctx context.Context,
	yookassaPaymentID string,
) (domain.Payment, error) {
	return r.scanPayment(r.db.QueryRow(ctx, sqlGetPaymentByYooKassaID, yookassaPaymentID))
}

func (r *PaymentRepo) InsertCoinsHistoryForPayment(
	ctx context.Context,
	userID, paymentID int64,
	coinsAmount int32,
	description string,
) error {
	referenceKey := "payment:" + strconv.FormatInt(paymentID, 10)

	_, err := r.db.Exec(
		ctx,
		sqlInsertCoinsHistoryForPayment,
		userID,
		coinsAmount,
		description,
		referenceKey,
	)
	if err != nil {
		return fmt.Errorf("insert vkino coins history for payment: %w", err)
	}

	return nil
}

func (r *PaymentRepo) TryRegisterWebhookEvent(
	ctx context.Context,
	yookassaPaymentID, event, payloadHash string,
) (bool, error) {
	var id int64

	err := r.db.QueryRow(
		ctx,
		sqlTryRegisterWebhookEvent,
		yookassaPaymentID,
		event,
		payloadHash,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("register webhook event: %w", err)
	}

	return true, nil
}

func (r *PaymentRepo) scanPayment(row pgx.Row) (domain.Payment, error) {
	var (
		payment           domain.Payment
		paymentMethod     string
		coinsSpent        sql.NullInt32
		yookassaPaymentID sql.NullString
		confirmationURL   sql.NullString
		paidAt            sql.NullTime
		productType       string
		status            string
	)

	err := row.Scan(
		&payment.ID,
		&payment.UserID,
		&productType,
		&payment.ProductRefID,
		&payment.Amount,
		&payment.Currency,
		&status,
		&paymentMethod,
		&coinsSpent,
		&yookassaPaymentID,
		&payment.IdempotencyKey,
		&confirmationURL,
		&paidAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Payment{}, domain.ErrPaymentNotFound
		}

		return domain.Payment{}, fmt.Errorf("scan payment: %w", err)
	}

	payment.ProductType = domain.ProductType(productType)
	payment.Status = domain.PaymentStatus(status)
	applyPaymentScanOptionals(&payment, paymentMethod, coinsSpent, yookassaPaymentID, confirmationURL, paidAt)

	return payment, nil
}

func applyPaymentScanOptionals(
	payment *domain.Payment,
	paymentMethod string,
	coinsSpent sql.NullInt32,
	yookassaPaymentID sql.NullString,
	confirmationURL sql.NullString,
	paidAt sql.NullTime,
) {
	payment.PaymentMethod = normalizePaymentMethodValue(paymentMethod)
	payment.CoinsSpent = nullableInt32Pointer(coinsSpent)
	payment.YooKassaPaymentID = nullableStringPointer(yookassaPaymentID)
	payment.ConfirmationURL = nullableStringPointer(confirmationURL)
	payment.PaidAt = nullableTimePointer(paidAt)
}

func normalizePaymentMethodValue(value string) domain.PaymentMethod {
	method := domain.PaymentMethod(value)
	if method == "" {
		return domain.PaymentMethodYooKassa
	}

	return method
}

func nullableInt32Pointer(value sql.NullInt32) *int32 {
	if !value.Valid {
		return nil
	}

	result := value.Int32

	return &result
}

func nullableStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String

	return &result
}

func nullableTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time

	return &result
}
