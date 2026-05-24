package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	if yookassaPaymentID.Valid {
		value := yookassaPaymentID.String
		payment.YooKassaPaymentID = &value
	}

	if confirmationURL.Valid {
		value := confirmationURL.String
		payment.ConfirmationURL = &value
	}

	if paidAt.Valid {
		value := paidAt.Time
		payment.PaidAt = &value
	}

	return payment, nil
}
