package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/usecase"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	yooKassaTestIP          = "185.71.76.10"
	tariffCodeLevel2        = "level_2"
	productTypeSubscription = "subscription"
	yooKassaStatusSucceeded = "succeeded"
)

var errUserServiceDown = errors.New("user service down")

func TestCreatePayment_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)
	activator := mocks.NewMockSubscriptionActivator(ctrl)

	payments.EXPECT().UserExists(gomock.Any(), int64(42)).Return(true, nil)
	payments.EXPECT().GetSubscriptionTariff(gomock.Any(), int64(2)).Return(domain.SubscriptionTariff{
		ID:                      2,
		Code:                    tariffCodeLevel2,
		Title:                   "Level 2",
		PriceMoney:              299,
		IsMoneyPaymentAvailable: true,
		DurationDays:            30,
		Level:                   2,
	}, nil)
	payments.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, payment domain.Payment) (domain.Payment, error) {
			require.Equal(t, int64(42), payment.UserID)
			require.Equal(t, domain.ProductTypeSubscription, payment.ProductType)
			require.Equal(t, int64(2), payment.ProductRefID)
			require.Equal(t, "299.00", payment.Amount)
			require.Equal(t, domain.PaymentStatusPending, payment.Status)

			payment.ID = 7

			return payment, nil
		},
	)
	yookassa.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).Return(repository.YooKassaPayment{
		ID:              "yk-7",
		Status:          "pending",
		ConfirmationURL: "https://yoomoney.ru/checkout/7",
	}, nil)
	payments.EXPECT().UpdatePaymentYooKassa(gomock.Any(), int64(7), "yk-7", "https://yoomoney.ru/checkout/7").Return(nil)

	u := usecase.New(payments, yookassa, activator, "http://localhost:3000/payments/return", true)

	result, err := u.CreatePayment(context.Background(), usecase.CreatePaymentInput{
		UserID:       42,
		ProductType:  productTypeSubscription,
		ProductRefID: 2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(7), result.PaymentID)
	require.Equal(t, "pending", result.Status)
	require.Equal(t, "https://yoomoney.ru/checkout/7", result.ConfirmationURL)
}

func TestCreatePayment_InvalidProductType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	u := usecase.New(
		mocks.NewMockPaymentRepo(ctrl),
		mocks.NewMockYooKassaClient(ctrl),
		mocks.NewMockSubscriptionActivator(ctrl),
		"http://localhost:3000/payments/return",
		true,
	)

	_, err := u.CreatePayment(context.Background(), usecase.CreatePaymentInput{
		UserID:       42,
		ProductType:  "coins",
		ProductRefID: 2,
	})
	require.ErrorIs(t, err, domain.ErrInvalidProductType)
}

func TestCreatePayment_TariffNotAvailable(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	payments.EXPECT().UserExists(gomock.Any(), int64(42)).Return(true, nil)
	payments.EXPECT().GetSubscriptionTariff(gomock.Any(), int64(2)).Return(domain.SubscriptionTariff{
		ID:                      2,
		Code:                    "free",
		IsMoneyPaymentAvailable: false,
	}, nil)

	u := usecase.New(
		payments,
		mocks.NewMockYooKassaClient(ctrl),
		mocks.NewMockSubscriptionActivator(ctrl),
		"http://localhost:3000/payments/return",
		true,
	)

	_, err := u.CreatePayment(context.Background(), usecase.CreatePaymentInput{
		UserID:       42,
		ProductType:  productTypeSubscription,
		ProductRefID: 2,
	})
	require.ErrorIs(t, err, domain.ErrTariffNotAvailable)
}

func TestCreatePayment_YooKassaFailureMarksPaymentCanceled(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)

	payments.EXPECT().UserExists(gomock.Any(), int64(42)).Return(true, nil)
	payments.EXPECT().GetSubscriptionTariff(gomock.Any(), int64(2)).Return(domain.SubscriptionTariff{
		ID:                      2,
		Code:                    tariffCodeLevel2,
		PriceMoney:              299,
		IsMoneyPaymentAvailable: true,
	}, nil)
	payments.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).Return(domain.Payment{ID: 9}, nil)
	yookassa.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).
		Return(repository.YooKassaPayment{}, domain.ErrYooKassaUnavailable)
	payments.EXPECT().UpdatePaymentStatus(gomock.Any(), int64(9), domain.PaymentStatusCanceled, nil).Return(nil)

	u := usecase.New(
		payments,
		yookassa,
		mocks.NewMockSubscriptionActivator(ctrl),
		"http://localhost:3000/payments/return",
		true,
	)

	_, err := u.CreatePayment(context.Background(), usecase.CreatePaymentInput{
		UserID:       42,
		ProductType:  productTypeSubscription,
		ProductRefID: 2,
	})
	require.ErrorIs(t, err, domain.ErrYooKassaUnavailable)
}

func TestGetPayment_AccessDenied(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	payments.EXPECT().GetPaymentByID(gomock.Any(), int64(5)).Return(domain.Payment{
		ID:     5,
		UserID: 99,
		Status: domain.PaymentStatusPending,
	}, nil)

	u := usecase.New(
		payments,
		mocks.NewMockYooKassaClient(ctrl),
		mocks.NewMockSubscriptionActivator(ctrl),
		"http://localhost:3000/payments/return",
		true,
	)

	_, err := u.GetPayment(context.Background(), 42, 5)
	require.ErrorIs(t, err, domain.ErrPaymentAccessDenied)
}

func TestGetPayment_SyncsSucceededFromYooKassa(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	yookassaPaymentID := "yk-sync-1"
	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)
	activator := mocks.NewMockSubscriptionActivator(ctrl)

	pendingPayment := domain.Payment{
		ID:                12,
		UserID:            42,
		ProductType:       domain.ProductTypeSubscription,
		ProductRefID:      2,
		Status:            domain.PaymentStatusPending,
		YooKassaPaymentID: &yookassaPaymentID,
	}
	succeededPayment := pendingPayment
	succeededPayment.Status = domain.PaymentStatusSucceeded

	gomock.InOrder(
		payments.EXPECT().GetPaymentByID(gomock.Any(), int64(12)).Return(pendingPayment, nil),
		yookassa.EXPECT().GetPayment(gomock.Any(), yookassaPaymentID).Return(repository.YooKassaPayment{
			ID:     yookassaPaymentID,
			Status: yooKassaStatusSucceeded,
			Paid:   true,
		}, nil),
		payments.EXPECT().UpdatePaymentStatus(
			gomock.Any(), int64(12), domain.PaymentStatusSucceeded, gomock.Any(),
		).Return(nil),
		activator.EXPECT().ActivateSubscription(gomock.Any(), int64(42), int64(2), int64(12)).Return(nil),
		payments.EXPECT().GetPaymentByID(gomock.Any(), int64(12)).Return(succeededPayment, nil),
	)

	u := usecase.New(payments, yookassa, activator, "http://localhost:3000/payments/return", true)

	payment, err := u.GetPayment(context.Background(), 42, 12)
	require.NoError(t, err)
	require.Equal(t, domain.PaymentStatusSucceeded, payment.Status)
}

func TestListMoneyTariffs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	payments.EXPECT().ListMoneyTariffs(gomock.Any()).Return([]domain.MoneyTariff{
		{ID: 2, Code: tariffCodeLevel2, PriceMoney: 299, Level: 2},
	}, nil)

	u := usecase.New(
		payments,
		mocks.NewMockYooKassaClient(ctrl),
		mocks.NewMockSubscriptionActivator(ctrl),
		"http://localhost:3000/payments/return",
		true,
	)

	tariffs, err := u.ListMoneyTariffs(context.Background())
	require.NoError(t, err)
	require.Len(t, tariffs, 1)
	require.Equal(t, tariffCodeLevel2, tariffs[0].Code)
}

func TestHandleYooKassaWebhook_InvalidIP(t *testing.T) {
	t.Parallel()

	u := usecase.New(
		mocks.NewMockPaymentRepo(gomock.NewController(t)),
		mocks.NewMockYooKassaClient(gomock.NewController(t)),
		mocks.NewMockSubscriptionActivator(gomock.NewController(t)),
		"http://localhost:3000/payments/return",
		true,
	)

	err := u.HandleYooKassaWebhook(context.Background(), []byte(`{}`), "127.0.0.1")
	require.ErrorIs(t, err, domain.ErrWebhookInvalidIP)
}

func TestHandleYooKassaWebhook_SucceededActivatesSubscription(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)
	activator := mocks.NewMockSubscriptionActivator(ctrl)

	body := []byte(`{
		"type":"notification",
		"event":"payment.succeeded",
		"object":{"id":"yk-42","status":"succeeded","paid":true}
	}`)

	payments.EXPECT().TryRegisterWebhookEvent(gomock.Any(), "yk-42", "payment.succeeded", gomock.Any()).Return(true, nil)
	yookassa.EXPECT().GetPayment(gomock.Any(), "yk-42").Return(repository.YooKassaPayment{
		ID:     "yk-42",
		Status: yooKassaStatusSucceeded,
		Paid:   true,
	}, nil)
	payments.EXPECT().GetPaymentByYooKassaID(gomock.Any(), "yk-42").Return(domain.Payment{
		ID:           11,
		UserID:       42,
		ProductType:  domain.ProductTypeSubscription,
		ProductRefID: 2,
		Status:       domain.PaymentStatusPending,
	}, nil)
	payments.EXPECT().UpdatePaymentStatus(gomock.Any(), int64(11), domain.PaymentStatusSucceeded, gomock.Any()).Return(nil)
	activator.EXPECT().ActivateSubscription(gomock.Any(), int64(42), int64(2), int64(11)).Return(nil)

	u := usecase.New(payments, yookassa, activator, "http://localhost:3000/payments/return", true)

	err := u.HandleYooKassaWebhook(context.Background(), body, yooKassaTestIP)
	require.NoError(t, err)
}

func TestHandleYooKassaWebhook_Canceled(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)

	body := []byte(`{
		"type":"notification",
		"event":"payment.canceled",
		"object":{"id":"yk-55","status":"canceled","paid":false}
	}`)

	payments.EXPECT().TryRegisterWebhookEvent(gomock.Any(), "yk-55", "payment.canceled", gomock.Any()).Return(true, nil)
	yookassa.EXPECT().GetPayment(gomock.Any(), "yk-55").Return(repository.YooKassaPayment{
		ID:     "yk-55",
		Status: "canceled",
	}, nil)
	payments.EXPECT().GetPaymentByYooKassaID(gomock.Any(), "yk-55").Return(domain.Payment{
		ID:     15,
		UserID: 42,
		Status: domain.PaymentStatusPending,
	}, nil)
	payments.EXPECT().UpdatePaymentStatus(gomock.Any(), int64(15), domain.PaymentStatusCanceled, nil).Return(nil)

	u := usecase.New(
		payments,
		yookassa,
		mocks.NewMockSubscriptionActivator(ctrl),
		"http://localhost:3000/payments/return",
		true,
	)

	err := u.HandleYooKassaWebhook(context.Background(), body, yooKassaTestIP)
	require.NoError(t, err)
}

func TestHandleYooKassaWebhook_DeduplicatedEventIsNoop(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	body := []byte(`{
		"type":"notification",
		"event":"payment.succeeded",
		"object":{"id":"yk-dup","status":"succeeded","paid":true}
	}`)

	payments.EXPECT().TryRegisterWebhookEvent(gomock.Any(), "yk-dup", "payment.succeeded", gomock.Any()).Return(false, nil)

	u := usecase.New(
		payments,
		mocks.NewMockYooKassaClient(ctrl),
		mocks.NewMockSubscriptionActivator(ctrl),
		"http://localhost:3000/payments/return",
		true,
	)

	err := u.HandleYooKassaWebhook(context.Background(), body, yooKassaTestIP)
	require.NoError(t, err)
}

func TestHandleYooKassaWebhook_ActivationFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payments := mocks.NewMockPaymentRepo(ctrl)
	yookassa := mocks.NewMockYooKassaClient(ctrl)
	activator := mocks.NewMockSubscriptionActivator(ctrl)

	body := []byte(`{
		"type":"notification",
		"event":"payment.succeeded",
		"object":{"id":"yk-fail","status":"succeeded","paid":true}
	}`)

	payments.EXPECT().TryRegisterWebhookEvent(gomock.Any(), "yk-fail", "payment.succeeded", gomock.Any()).Return(true, nil)
	yookassa.EXPECT().GetPayment(gomock.Any(), "yk-fail").Return(repository.YooKassaPayment{
		ID:     "yk-fail",
		Status: yooKassaStatusSucceeded,
		Paid:   true,
	}, nil)
	payments.EXPECT().GetPaymentByYooKassaID(gomock.Any(), "yk-fail").Return(domain.Payment{
		ID:           20,
		UserID:       42,
		ProductType:  domain.ProductTypeSubscription,
		ProductRefID: 2,
		Status:       domain.PaymentStatusPending,
	}, nil)
	payments.EXPECT().UpdatePaymentStatus(gomock.Any(), int64(20), domain.PaymentStatusSucceeded, gomock.Any()).Return(nil)
	activator.EXPECT().ActivateSubscription(gomock.Any(), int64(42), int64(2), int64(20)).Return(errUserServiceDown)

	u := usecase.New(
		payments,
		yookassa,
		activator,
		"http://localhost:3000/payments/return",
		true,
	)

	err := u.HandleYooKassaWebhook(context.Background(), body, yooKassaTestIP)
	require.ErrorIs(t, err, domain.ErrInternal)
}
