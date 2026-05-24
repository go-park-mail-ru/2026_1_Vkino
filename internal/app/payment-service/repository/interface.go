package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
)

//go:generate mockgen -source=./interface.go -destination=./mocks/repository_mock.go -package=mocks

type PaymentRepo interface {
	UserExists(ctx context.Context, userID int64) (bool, error)
	GetSubscriptionTariff(ctx context.Context, tariffID int64) (domain.SubscriptionTariff, error)
	ListMoneyTariffs(ctx context.Context) ([]domain.MoneyTariff, error)
	CreatePayment(ctx context.Context, payment domain.Payment) (domain.Payment, error)
	UpdatePaymentYooKassa(
		ctx context.Context,
		paymentID int64,
		yookassaPaymentID, confirmationURL string,
	) error
	UpdatePaymentStatus(ctx context.Context, paymentID int64, status domain.PaymentStatus, paidAt *time.Time) error
	GetPaymentByID(ctx context.Context, paymentID int64) (domain.Payment, error)
	GetPaymentByYooKassaID(ctx context.Context, yookassaPaymentID string) (domain.Payment, error)
	TryRegisterWebhookEvent(
		ctx context.Context,
		yookassaPaymentID, event, payloadHash string,
	) (bool, error)
}

type YooKassaCreateRequest struct {
	AmountValue      string
	AmountCurrency   string
	Capture          bool
	ReturnURL        string
	Description      string
	IdempotencyKey   string
	Metadata         map[string]string
}

type YooKassaPayment struct {
	ID               string
	Status           string
	Paid             bool
	ConfirmationURL  string
}

type YooKassaClient interface {
	CreatePayment(ctx context.Context, req YooKassaCreateRequest) (YooKassaPayment, error)
	GetPayment(ctx context.Context, paymentID string) (YooKassaPayment, error)
}

type SubscriptionActivator interface {
	ActivateSubscription(ctx context.Context, userID, tariffID, paymentID int64) error
}
