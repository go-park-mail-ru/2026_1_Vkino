package grpc

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserSubscriptionCoinsBuyer struct {
	client userv1.UserServiceClient
}

func NewUserSubscriptionCoinsBuyer(client userv1.UserServiceClient) *UserSubscriptionCoinsBuyer {
	return &UserSubscriptionCoinsBuyer{client: client}
}

func (b *UserSubscriptionCoinsBuyer) BuySubscriptionWithVKinoCoins(
	ctx context.Context,
	userID, tariffID int64,
) (domain.CoinsSubscriptionPurchase, error) {
	resp, err := b.client.BuySubscriptionWithVKinoCoins(ctx, &userv1.BuySubscriptionWithVKinoCoinsRequest{
		UserId:   userID,
		TariffId: tariffID,
	})
	if err != nil {
		return domain.CoinsSubscriptionPurchase{}, mapBuySubscriptionWithVKinoCoinsError(err)
	}

	return domain.CoinsSubscriptionPurchase{
		PaymentID:         resp.GetPaymentId(),
		CoinsSpent:        resp.GetCoinsSpent(),
		VKinoCoinsBalance: resp.GetVkinoCoinsBalance(),
	}, nil
}

func mapBuySubscriptionWithVKinoCoinsError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("buy subscription with vkino coins: %w", err)
	}

	if mappedErr, ok := buySubscriptionWithVKinoCoinsGRPCErrors[buySubscriptionWithVKinoCoinsErrorKey{
		code:    st.Code(),
		message: st.Message(),
	}]; ok {
		return mappedErr
	}

	return fmt.Errorf("buy subscription with vkino coins: %w", err)
}

type buySubscriptionWithVKinoCoinsErrorKey struct {
	code    codes.Code
	message string
}

var buySubscriptionWithVKinoCoinsGRPCErrors = map[buySubscriptionWithVKinoCoinsErrorKey]error{
	{code: codes.Unauthenticated, message: "unauthorized"}:                domain.ErrInvalidToken,
	{code: codes.NotFound, message: "subscription tariff not found"}:      domain.ErrTariffNotFound,
	{code: codes.FailedPrecondition, message: "insufficient vkino coins"}: domain.ErrInsufficientVKinoCoins,
	{
		code:    codes.FailedPrecondition,
		message: "tariff is not available for vkino coins payment",
	}: domain.ErrTariffNotAvailableForCoins,
}
