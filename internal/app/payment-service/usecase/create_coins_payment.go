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

func (u *Usecase) validateCoinsPaymentInput(input CreatePaymentInput) error {
	if input.UserID <= 0 {
		return domain.ErrInvalidToken
	}

	productType := domain.ProductType(strings.TrimSpace(input.ProductType))
	if productType != domain.ProductTypeCoins {
		return domain.ErrInvalidProductType
	}

	if input.ProductRefID <= 0 {
		return domain.ErrInvalidProductRef
	}

	if normalizePaymentMethod(input.PaymentMethod) != domain.PaymentMethodYooKassa {
		return domain.ErrInvalidPaymentMethod
	}

	return nil
}

func (u *Usecase) createYooKassaCoinsPayment(
	ctx context.Context,
	input CreatePaymentInput,
) (CreatePaymentResult, error) {
	if err := u.validateCoinsPaymentInput(input); err != nil {
		return CreatePaymentResult{}, err
	}

	if err := u.ensurePaymentUserExists(ctx, input.UserID); err != nil {
		return CreatePaymentResult{}, err
	}

	pack, err := u.loadCoinsPack(ctx, input.ProductRefID)
	if err != nil {
		return CreatePaymentResult{}, err
	}

	payment, amountValue, err := u.createPendingCoinsPayment(ctx, input, pack)
	if err != nil {
		return CreatePaymentResult{}, err
	}

	return u.confirmCoinsPaymentWithYooKassa(ctx, input, pack, payment, amountValue)
}

func (u *Usecase) confirmCoinsPaymentWithYooKassa(
	ctx context.Context,
	input CreatePaymentInput,
	pack domain.CoinsPack,
	payment domain.Payment,
	amountValue string,
) (CreatePaymentResult, error) {
	ykPayment, err := u.requestYooKassaPaymentForCoinsPack(ctx, input, pack, payment, amountValue)
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

func (u *Usecase) loadCoinsPack(ctx context.Context, packID int64) (domain.CoinsPack, error) {
	pack, err := u.payments.GetCoinsPack(ctx, packID)
	if err != nil {
		return domain.CoinsPack{}, err
	}

	if pack.PriceMoney <= 0 || pack.CoinsAmount <= 0 {
		return domain.CoinsPack{}, domain.ErrCoinsPackNotFound
	}

	return pack, nil
}

func (u *Usecase) createPendingCoinsPayment(
	ctx context.Context,
	input CreatePaymentInput,
	pack domain.CoinsPack,
) (domain.Payment, string, error) {
	amountValue := fmt.Sprintf("%d.00", pack.PriceMoney)

	payment, err := u.payments.CreatePayment(ctx, domain.Payment{
		UserID:         input.UserID,
		ProductType:    domain.ProductTypeCoins,
		ProductRefID:   pack.ID,
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

func (u *Usecase) requestYooKassaPaymentForCoinsPack(
	ctx context.Context,
	input CreatePaymentInput,
	pack domain.CoinsPack,
	payment domain.Payment,
	amountValue string,
) (repository.YooKassaPayment, error) {
	description := "VKino coins " + pack.Title

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
			"product_type": string(domain.ProductTypeCoins),
			"pack_id":      strconv.FormatInt(pack.ID, 10),
		},
	})
}

func (u *Usecase) creditCoinsForPayment(ctx context.Context, payment domain.Payment) error {
	pack, err := u.loadCoinsPack(ctx, payment.ProductRefID)
	if err != nil {
		return err
	}

	description := "Покупка " + pack.Title

	if err = u.payments.InsertCoinsHistoryForPayment(
		ctx,
		payment.UserID,
		payment.ID,
		pack.CoinsAmount,
		description,
	); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return nil
}
