package grpc

import (
	"context"
	"fmt"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
)

type UserSubscriptionActivator struct {
	client userv1.UserServiceClient
}

func NewUserSubscriptionActivator(client userv1.UserServiceClient) *UserSubscriptionActivator {
	return &UserSubscriptionActivator{client: client}
}

func (a *UserSubscriptionActivator) ActivateSubscription(
	ctx context.Context,
	userID, tariffID, paymentID int64,
) error {
	_, err := a.client.ActivateSubscription(ctx, &userv1.ActivateSubscriptionRequest{
		UserId:    userID,
		TariffId:  tariffID,
		PaymentId: paymentID,
	})
	if err != nil {
		return fmt.Errorf("activate subscription: %w", err)
	}

	return nil
}
