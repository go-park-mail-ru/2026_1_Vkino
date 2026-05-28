package usecase_test

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/usecase"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	productTypeCoins = "coins"
	packCodeStarter  = "pack_50"
)

func TestCreatePayment_Coins_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)

	payments.EXPECT().UserExists(gomock.Any(), int64(42)).Return(true, nil)
	payments.EXPECT().GetCoinsPack(gomock.Any(), int64(1)).Return(domain.CoinsPack{
		ID:          1,
		Code:        packCodeStarter,
		Title:       "50 VKino coins",
		CoinsAmount: 50,
		PriceMoney:  99,
	}, nil)
	payments.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, payment domain.Payment) (domain.Payment, error) {
			require.Equal(t, int64(42), payment.UserID)
			require.Equal(t, domain.ProductTypeCoins, payment.ProductType)
			require.Equal(t, int64(1), payment.ProductRefID)
			require.Equal(t, "99.00", payment.Amount)
			require.Equal(t, domain.PaymentStatusPending, payment.Status)
			require.Equal(t, domain.PaymentMethodYooKassa, payment.PaymentMethod)

			payment.ID = 15

			return payment, nil
		},
	)
	yookassa.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).Return(repository.YooKassaPayment{
		ID:              "yk-coins-15",
		Status:          "pending",
		ConfirmationURL: "https://yoomoney.ru/checkout/coins-15",
	}, nil)
	payments.EXPECT().UpdatePaymentYooKassa(
		gomock.Any(),
		int64(15),
		"yk-coins-15",
		"https://yoomoney.ru/checkout/coins-15",
	).Return(nil)

	u := usecase.New(payments, yookassa, nil, nil, "http://localhost:3000/payments/return", true)

	result, err := u.CreatePayment(context.Background(), usecase.CreatePaymentInput{
		UserID:       42,
		ProductType:  productTypeCoins,
		ProductRefID: 1,
	})
	require.NoError(t, err)
	require.Equal(t, int64(15), result.PaymentID)
	require.Equal(t, "pending", result.Status)
	require.Equal(t, "https://yoomoney.ru/checkout/coins-15", result.ConfirmationURL)
	require.Equal(t, "yookassa", result.PaymentMethod)
}

func TestCreatePayment_Coins_PackNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)

	payments.EXPECT().UserExists(gomock.Any(), int64(42)).Return(true, nil)
	payments.EXPECT().GetCoinsPack(gomock.Any(), int64(999)).Return(domain.CoinsPack{}, domain.ErrCoinsPackNotFound)

	u := usecase.New(
		payments,
		mocks.NewMockYooKassaClient(ctrl),
		nil,
		nil,
		"http://localhost:3000/payments/return",
		true,
	)

	_, err := u.CreatePayment(context.Background(), usecase.CreatePaymentInput{
		UserID:       42,
		ProductType:  productTypeCoins,
		ProductRefID: 999,
	})
	require.ErrorIs(t, err, domain.ErrCoinsPackNotFound)
}

func TestHandleYooKassaWebhook_Coins_InsertsHistory(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)

	body := []byte(`{
		"type":"notification",
		"event":"payment.succeeded",
		"object":{"id":"yk-coins-20","status":"succeeded","paid":true}
	}`)

	payments.EXPECT().TryRegisterWebhookEvent(gomock.Any(), "yk-coins-20", "payment.succeeded", gomock.Any()).
		Return(true, nil)
	yookassa.EXPECT().GetPayment(gomock.Any(), "yk-coins-20").Return(repository.YooKassaPayment{
		ID:     "yk-coins-20",
		Status: yooKassaStatusSucceeded,
		Paid:   true,
	}, nil)
	payments.EXPECT().GetPaymentByYooKassaID(gomock.Any(), "yk-coins-20").Return(domain.Payment{
		ID:            20,
		UserID:        42,
		ProductType:   domain.ProductTypeCoins,
		ProductRefID:  2,
		Status:        domain.PaymentStatusPending,
		PaymentMethod: domain.PaymentMethodYooKassa,
	}, nil)
	payments.EXPECT().UpdatePaymentStatus(gomock.Any(), int64(20), domain.PaymentStatusSucceeded, gomock.Any()).
		Return(nil)
	payments.EXPECT().GetCoinsPack(gomock.Any(), int64(2)).Return(domain.CoinsPack{
		ID:          2,
		Code:        "pack_150",
		Title:       "150 VKino coins",
		CoinsAmount: 150,
		PriceMoney:  249,
	}, nil)
	payments.EXPECT().InsertCoinsHistoryForPayment(
		gomock.Any(),
		int64(42),
		int64(20),
		int32(150),
		"Покупка 150 VKino coins",
	).Return(nil)

	u := usecase.New(payments, yookassa, nil, nil, "http://localhost:3000/payments/return", true)

	err := u.HandleYooKassaWebhook(context.Background(), body, yooKassaTestIP)
	require.NoError(t, err)
}

func TestHandleYooKassaWebhook_Coins_AlreadySucceededSkipsHistory(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)

	body := []byte(`{
		"type":"notification",
		"event":"payment.succeeded",
		"object":{"id":"yk-coins-30","status":"succeeded","paid":true}
	}`)

	payments.EXPECT().TryRegisterWebhookEvent(gomock.Any(), "yk-coins-30", "payment.succeeded", gomock.Any()).
		Return(true, nil)
	yookassa.EXPECT().GetPayment(gomock.Any(), "yk-coins-30").Return(repository.YooKassaPayment{
		ID:     "yk-coins-30",
		Status: yooKassaStatusSucceeded,
		Paid:   true,
	}, nil)
	payments.EXPECT().GetPaymentByYooKassaID(gomock.Any(), "yk-coins-30").Return(domain.Payment{
		ID:            30,
		UserID:        42,
		ProductType:   domain.ProductTypeCoins,
		ProductRefID:  1,
		Status:        domain.PaymentStatusSucceeded,
		PaymentMethod: domain.PaymentMethodYooKassa,
	}, nil)

	u := usecase.New(payments, yookassa, nil, nil, "http://localhost:3000/payments/return", true)

	err := u.HandleYooKassaWebhook(context.Background(), body, yooKassaTestIP)
	require.NoError(t, err)
}
