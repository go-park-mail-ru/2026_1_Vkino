package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository"
	yookassapkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/yookassa"
	"github.com/google/uuid"
)

type Usecase struct {
	payments     repository.PaymentRepo
	yookassa     repository.YooKassaClient
	activator    repository.SubscriptionActivator
	returnURL    string
	capture      bool
	now          func() time.Time
}

func New(
	payments repository.PaymentRepo,
	yookassa repository.YooKassaClient,
	activator repository.SubscriptionActivator,
	returnURL string,
	capture bool,
) *Usecase {
	return &Usecase{
		payments:  payments,
		yookassa:  yookassa,
		activator: activator,
		returnURL: returnURL,
		capture:   capture,
		now:       time.Now,
	}
}

type CreatePaymentInput struct {
	UserID       int64
	ProductType  string
	ProductRefID int64
}

type CreatePaymentResult struct {
	PaymentID        int64
	Status           string
	ConfirmationURL  string
}

func (u *Usecase) CreatePayment(ctx context.Context, input CreatePaymentInput) (CreatePaymentResult, error) {
	if input.UserID <= 0 {
		return CreatePaymentResult{}, domain.ErrInvalidToken
	}

	productType := domain.ProductType(strings.TrimSpace(input.ProductType))
	if productType != domain.ProductTypeSubscription {
		return CreatePaymentResult{}, domain.ErrInvalidProductType
	}

	if input.ProductRefID <= 0 {
		return CreatePaymentResult{}, domain.ErrInvalidProductRef
	}

	exists, err := u.payments.UserExists(ctx, input.UserID)
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	if !exists {
		return CreatePaymentResult{}, domain.ErrInvalidToken
	}

	tariff, err := u.payments.GetSubscriptionTariff(ctx, input.ProductRefID)
	if err != nil {
		return CreatePaymentResult{}, err
	}

	if !tariff.IsMoneyPaymentAvailable || tariff.PriceMoney <= 0 {
		return CreatePaymentResult{}, domain.ErrTariffNotAvailable
	}

	amountValue := fmt.Sprintf("%d.00", tariff.PriceMoney)

	payment, err := u.payments.CreatePayment(ctx, domain.Payment{
		UserID:         input.UserID,
		ProductType:    productType,
		ProductRefID:   input.ProductRefID,
		Amount:         amountValue,
		Currency:       "RUB",
		Status:         domain.PaymentStatusPending,
		IdempotencyKey: uuid.NewString(),
	})
	if err != nil {
		return CreatePaymentResult{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	description := fmt.Sprintf("VKino subscription %s", tariff.Code)

	ykPayment, err := u.yookassa.CreatePayment(ctx, repository.YooKassaCreateRequest{
		AmountValue:    amountValue,
		AmountCurrency: "RUB",
		Capture:        u.capture,
		ReturnURL:      fmt.Sprintf("%s?payment_id=%d", strings.TrimRight(u.returnURL, "/"), payment.ID),
		Description:    description,
		IdempotencyKey: payment.IdempotencyKey,
		Metadata: map[string]string{
			"payment_id":   strconv.FormatInt(payment.ID, 10),
			"user_id":      strconv.FormatInt(input.UserID, 10),
			"product_type": string(productType),
			"tariff_id":    strconv.FormatInt(tariff.ID, 10),
		},
	})
	if err != nil {
		_ = u.payments.UpdatePaymentStatus(ctx, payment.ID, domain.PaymentStatusCanceled, nil)

		return CreatePaymentResult{}, err
	}

	if err = u.payments.UpdatePaymentYooKassa(
		ctx,
		payment.ID,
		ykPayment.ID,
		ykPayment.ConfirmationURL,
	); err != nil {
		return CreatePaymentResult{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return CreatePaymentResult{
		PaymentID:       payment.ID,
		Status:          string(domain.PaymentStatusPending),
		ConfirmationURL: ykPayment.ConfirmationURL,
	}, nil
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
		if syncErr := u.syncPendingPaymentFromYooKassa(ctx, payment); syncErr == nil {
			payment, err = u.payments.GetPaymentByID(ctx, paymentID)
			if err != nil {
				return domain.Payment{}, err
			}
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

	var notification webhookNotification
	if err := json.Unmarshal(body, &notification); err != nil {
		return domain.ErrWebhookInvalidPayload
	}

	if notification.Type != "notification" || notification.Object.ID == "" {
		return domain.ErrWebhookInvalidPayload
	}

	payloadHash := sha256.Sum256(body)
	hash := hex.EncodeToString(payloadHash[:])

	registered, err := u.payments.TryRegisterWebhookEvent(ctx, notification.Object.ID, notification.Event, hash)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	if !registered {
		return nil
	}

	ykPayment, err := u.yookassa.GetPayment(ctx, notification.Object.ID)
	if err != nil {
		return err
	}

	payment, err := u.payments.GetPaymentByYooKassaID(ctx, notification.Object.ID)
	if err != nil {
		return err
	}

	switch notification.Event {
	case "payment.succeeded":
		if ykPayment.Status != "succeeded" {
			return domain.ErrWebhookInvalidPayload
		}

		return u.finalizeSucceededPayment(ctx, payment)

	case "payment.canceled":
		return u.finalizeCanceledPayment(ctx, payment)
	}

	return nil
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
