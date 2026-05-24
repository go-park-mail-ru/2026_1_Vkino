package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository"
	"github.com/google/uuid"
)

func (u *Usecase) validateCreatePaymentInput(
	ctx context.Context,
	input CreatePaymentInput,
) (domain.ProductType, domain.SubscriptionTariff, error) {
	productType, err := u.validateCreatePaymentProduct(input)
	if err != nil {
		return "", domain.SubscriptionTariff{}, err
	}

	if ensureErr := u.ensurePaymentUserExists(ctx, input.UserID); ensureErr != nil {
		return "", domain.SubscriptionTariff{}, ensureErr
	}

	tariff, err := u.loadMoneyTariff(ctx, input.ProductRefID)
	if err != nil {
		return "", domain.SubscriptionTariff{}, err
	}

	return productType, tariff, nil
}

func (u *Usecase) validateCreatePaymentProduct(input CreatePaymentInput) (domain.ProductType, error) {
	if input.UserID <= 0 {
		return "", domain.ErrInvalidToken
	}

	productType := domain.ProductType(strings.TrimSpace(input.ProductType))
	if productType != domain.ProductTypeSubscription {
		return "", domain.ErrInvalidProductType
	}

	if input.ProductRefID <= 0 {
		return "", domain.ErrInvalidProductRef
	}

	return productType, nil
}

func (u *Usecase) ensurePaymentUserExists(ctx context.Context, userID int64) error {
	exists, err := u.payments.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	if !exists {
		return domain.ErrInvalidToken
	}

	return nil
}

func (u *Usecase) loadMoneyTariff(ctx context.Context, productRefID int64) (domain.SubscriptionTariff, error) {
	tariff, err := u.payments.GetSubscriptionTariff(ctx, productRefID)
	if err != nil {
		return domain.SubscriptionTariff{}, err
	}

	if !tariff.IsMoneyPaymentAvailable || tariff.PriceMoney <= 0 {
		return domain.SubscriptionTariff{}, domain.ErrTariffNotAvailable
	}

	return tariff, nil
}

func (u *Usecase) createPendingPayment(
	ctx context.Context,
	input CreatePaymentInput,
	productType domain.ProductType,
	tariff domain.SubscriptionTariff,
) (domain.Payment, string, error) {
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
		return domain.Payment{}, "", fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return payment, amountValue, nil
}

func (u *Usecase) createYooKassaPayment(
	ctx context.Context,
	input CreatePaymentInput,
	productType domain.ProductType,
	tariff domain.SubscriptionTariff,
	payment domain.Payment,
	amountValue string,
) (repository.YooKassaPayment, error) {
	description := "VKino subscription " + tariff.Code

	return u.yookassa.CreatePayment(ctx, repository.YooKassaCreateRequest{
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
}
