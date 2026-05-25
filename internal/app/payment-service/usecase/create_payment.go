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

func (u *Usecase) validateSubscriptionPaymentInput(input CreatePaymentInput) (domain.ProductType, error) {
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

func normalizePaymentMethod(raw string) domain.PaymentMethod {
	method := domain.PaymentMethod(strings.TrimSpace(raw))
	if method == "" {
		return domain.PaymentMethodYooKassa
	}

	return method
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

func (u *Usecase) createYooKassaSubscriptionPayment(
	ctx context.Context,
	input CreatePaymentInput,
) (CreatePaymentResult, error) {
	productType, tariff, err := u.prepareYooKassaSubscriptionPayment(ctx, input)
	if err != nil {
		return CreatePaymentResult{}, err
	}

	payment, amountValue, err := u.createPendingPayment(ctx, input, productType, tariff)
	if err != nil {
		return CreatePaymentResult{}, err
	}

	ykPayment, err := u.createYooKassaPayment(ctx, input, productType, tariff, payment, amountValue)
	if err != nil {
		if cancelErr := u.cancelPaymentAfterYooKassaFailure(ctx, payment.ID); cancelErr != nil {
			return CreatePaymentResult{}, cancelErr
		}

		return CreatePaymentResult{}, err
	}

	if err = u.linkYooKassaPayment(ctx, payment.ID, ykPayment); err != nil {
		return CreatePaymentResult{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return CreatePaymentResult{
		PaymentID:       payment.ID,
		Status:          string(domain.PaymentStatusPending),
		ConfirmationURL: ykPayment.ConfirmationURL,
		PaymentMethod:   string(domain.PaymentMethodYooKassa),
	}, nil
}

func (u *Usecase) prepareYooKassaSubscriptionPayment(
	ctx context.Context,
	input CreatePaymentInput,
) (domain.ProductType, domain.SubscriptionTariff, error) {
	productType, err := u.validateSubscriptionPaymentInput(input)
	if err != nil {
		return "", domain.SubscriptionTariff{}, err
	}

	if err = u.ensurePaymentUserExists(ctx, input.UserID); err != nil {
		return "", domain.SubscriptionTariff{}, err
	}

	tariff, err := u.loadMoneyTariff(ctx, input.ProductRefID)
	if err != nil {
		return "", domain.SubscriptionTariff{}, err
	}

	return productType, tariff, nil
}

func (u *Usecase) cancelPaymentAfterYooKassaFailure(ctx context.Context, paymentID int64) error {
	if err := u.payments.UpdatePaymentStatus(ctx, paymentID, domain.PaymentStatusCanceled, nil); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return nil
}

func (u *Usecase) linkYooKassaPayment(
	ctx context.Context,
	paymentID int64,
	ykPayment repository.YooKassaPayment,
) error {
	return u.payments.UpdatePaymentYooKassa(
		ctx,
		paymentID,
		ykPayment.ID,
		ykPayment.ConfirmationURL,
	)
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

func (u *Usecase) loadCoinsTariff(ctx context.Context, productRefID int64) (domain.SubscriptionTariff, error) {
	tariff, err := u.payments.GetSubscriptionTariff(ctx, productRefID)
	if err != nil {
		return domain.SubscriptionTariff{}, err
	}

	if !tariff.IsCoinsPaymentAvailable || tariff.PriceVKinoCoins <= 0 {
		return domain.SubscriptionTariff{}, domain.ErrTariffNotAvailableForCoins
	}

	return tariff, nil
}

func (u *Usecase) createVKinoCoinsSubscriptionPayment(
	ctx context.Context,
	input CreatePaymentInput,
) (CreatePaymentResult, error) {
	if _, err := u.validateSubscriptionPaymentInput(input); err != nil {
		return CreatePaymentResult{}, err
	}

	if err := u.ensurePaymentUserExists(ctx, input.UserID); err != nil {
		return CreatePaymentResult{}, err
	}

	if _, err := u.loadCoinsTariff(ctx, input.ProductRefID); err != nil {
		return CreatePaymentResult{}, err
	}

	purchase, err := u.coinsBuyer.BuySubscriptionWithVKinoCoins(ctx, input.UserID, input.ProductRefID)
	if err != nil {
		return CreatePaymentResult{}, err
	}

	coinsSpent := purchase.CoinsSpent
	balance := purchase.VKinoCoinsBalance

	return CreatePaymentResult{
		PaymentID:         purchase.PaymentID,
		Status:            string(domain.PaymentStatusSucceeded),
		ConfirmationURL:   "",
		PaymentMethod:     string(domain.PaymentMethodVKinoCoins),
		CoinsSpent:        &coinsSpent,
		VKinoCoinsBalance: &balance,
	}, nil
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
		PaymentMethod:  domain.PaymentMethodYooKassa,
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
