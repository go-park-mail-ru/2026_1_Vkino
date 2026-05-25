package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository"
	yookassapkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/yookassa"
)

type Usecase struct {
	payments   repository.PaymentRepo
	yookassa   repository.YooKassaClient
	coinsBuyer repository.SubscriptionCoinsBuyer
	activator  repository.SubscriptionActivator
	returnURL  string
	capture    bool
	now        func() time.Time
}

func New(
	payments repository.PaymentRepo,
	yookassa repository.YooKassaClient,
	coinsBuyer repository.SubscriptionCoinsBuyer,
	activator repository.SubscriptionActivator,
	returnURL string,
	capture bool,
) *Usecase {
	return &Usecase{
		payments:   payments,
		yookassa:   yookassa,
		coinsBuyer: coinsBuyer,
		activator:  activator,
		returnURL:  returnURL,
		capture:    capture,
		now:        time.Now,
	}
}

type CreatePaymentInput struct {
	UserID        int64
	ProductType   string
	ProductRefID  int64
	PaymentMethod string
}

type CreatePaymentResult struct {
	PaymentID         int64
	Status            string
	ConfirmationURL   string
	PaymentMethod     string
	CoinsSpent        *int32
	VKinoCoinsBalance *int32
}

func (u *Usecase) CreatePayment(ctx context.Context, input CreatePaymentInput) (CreatePaymentResult, error) {
	switch normalizePaymentMethod(input.PaymentMethod) {
	case domain.PaymentMethodYooKassa:
		return u.createYooKassaSubscriptionPayment(ctx, input)
	case domain.PaymentMethodVKinoCoins:
		return u.createVKinoCoinsSubscriptionPayment(ctx, input)
	default:
		return CreatePaymentResult{}, domain.ErrInvalidPaymentMethod
	}
}

func (u *Usecase) GetPayment(ctx context.Context, userID, paymentID int64) (domain.Payment, error) {
	if userID <= 0 {
		return domain.Payment{}, domain.ErrInvalidToken
	}

	payment, err := u.payments.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return domain.Payment{}, err
	}

	if payment.UserID != userID {
		return domain.Payment{}, domain.ErrPaymentAccessDenied
	}

	if payment.Status == domain.PaymentStatusPending {
		payment, err = u.refreshPendingPayment(ctx, paymentID, payment)
		if err != nil {
			return domain.Payment{}, err
		}
	}

	return payment, nil
}

func (u *Usecase) ListMoneyTariffs(ctx context.Context) ([]domain.MoneyTariff, error) {
	tariffs, err := u.payments.ListMoneyTariffs(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return tariffs, nil
}

type webhookNotification struct {
	Type   string `json:"type"`
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Paid   bool   `json:"paid"`
	} `json:"object"`
}

func (u *Usecase) HandleYooKassaWebhook(ctx context.Context, body []byte, clientIP string) error {
	if !yookassapkg.IsYooKassaIP(clientIP) {
		return domain.ErrWebhookInvalidIP
	}

	notification, err := parseWebhookNotification(body)
	if err != nil {
		return err
	}

	registered, err := u.registerWebhookEvent(ctx, notification, body)
	if err != nil {
		return err
	}

	if !registered {
		return nil
	}

	return u.dispatchWebhookEvent(ctx, notification)
}

func (u *Usecase) syncPendingPaymentFromYooKassa(ctx context.Context, payment domain.Payment) error {
	if payment.YooKassaPaymentID == nil {
		return nil
	}

	yookassaPaymentID := strings.TrimSpace(*payment.YooKassaPaymentID)
	if yookassaPaymentID == "" {
		return nil
	}

	ykPayment, err := u.yookassa.GetPayment(ctx, yookassaPaymentID)
	if err != nil {
		return err
	}

	switch ykPayment.Status {
	case "succeeded":
		if !ykPayment.Paid {
			return nil
		}

		return u.finalizeSucceededPayment(ctx, payment)
	case "canceled":
		return u.finalizeCanceledPayment(ctx, payment)
	default:
		return nil
	}
}

func (u *Usecase) finalizeSucceededPayment(ctx context.Context, payment domain.Payment) error {
	if payment.Status == domain.PaymentStatusSucceeded {
		return nil
	}

	now := u.now()

	if err := u.payments.UpdatePaymentStatus(ctx, payment.ID, domain.PaymentStatusSucceeded, &now); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	if payment.ProductType != domain.ProductTypeSubscription {
		return nil
	}

	if err := u.activator.ActivateSubscription(
		ctx,
		payment.UserID,
		payment.ProductRefID,
		payment.ID,
	); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return nil
}

func (u *Usecase) finalizeCanceledPayment(ctx context.Context, payment domain.Payment) error {
	if payment.Status == domain.PaymentStatusSucceeded {
		return nil
	}

	if err := u.payments.UpdatePaymentStatus(ctx, payment.ID, domain.PaymentStatusCanceled, nil); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return nil
}
