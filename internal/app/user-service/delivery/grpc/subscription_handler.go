package grpc

import (
	"context"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
)

func (s *Server) ActivateSubscription(
	ctx context.Context,
	req *userv1.ActivateSubscriptionRequest,
) (*userv1.ActivateSubscriptionResponse, error) {
	info, err := s.usecase.ActivateSubscription(
		ctx,
		req.GetUserId(),
		req.GetTariffId(),
		req.GetPaymentId(),
	)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.ActivateSubscriptionResponse{
		Subscription: &userv1.SubscriptionInfo{
			Id:    info.ID,
			Code:  info.Code,
			Name:  info.Name,
			Level: info.Level,
		},
	}

	if info.ActiveUntil != nil {
		resp.Subscription.ActiveUntil = info.ActiveUntil
	}

	return resp, nil
}
