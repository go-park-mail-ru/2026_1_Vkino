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

func (s *Server) BuySubscriptionWithVKinoCoins(
	ctx context.Context,
	req *userv1.BuySubscriptionWithVKinoCoinsRequest,
) (*userv1.BuySubscriptionWithVKinoCoinsResponse, error) {
	purchase, err := s.usecase.BuySubscriptionWithVKinoCoins(ctx, req.GetUserId(), req.GetTariffId())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.BuySubscriptionWithVKinoCoinsResponse{
		PaymentId:         purchase.PaymentID,
		CoinsSpent:        purchase.CoinsSpent,
		VkinoCoinsBalance: purchase.VKinoCoinsBalance,
		Subscription: &userv1.SubscriptionInfo{
			Id:    purchase.Subscription.ID,
			Code:  purchase.Subscription.Code,
			Name:  purchase.Subscription.Name,
			Level: purchase.Subscription.Level,
		},
	}

	if purchase.Subscription.ActiveUntil != nil {
		resp.Subscription.ActiveUntil = purchase.Subscription.ActiveUntil
	}

	return resp, nil
}

func (s *Server) SpendVKinoCoins(
	ctx context.Context,
	req *userv1.SpendVKinoCoinsRequest,
) (*userv1.SpendVKinoCoinsResponse, error) {
	spend, err := s.usecase.SpendVKinoCoins(
		ctx,
		req.GetUserId(),
		req.GetCoinsAmount(),
		req.GetOperationType(),
		req.GetDescription(),
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.SpendVKinoCoinsResponse{
		CoinsSpent:        spend.CoinsSpent,
		VkinoCoinsBalance: spend.VKinoCoinsBalance,
	}, nil
}

func (s *Server) GrantVKinoCoins(
	ctx context.Context,
	req *userv1.GrantVKinoCoinsRequest,
) (*userv1.GrantVKinoCoinsResponse, error) {
	grant, err := s.usecase.GrantVKinoCoins(
		ctx,
		req.GetUserId(),
		req.GetCoinsAmount(),
		req.GetOperationType(),
		req.GetDescription(),
		req.GetReferenceKey(),
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.GrantVKinoCoinsResponse{
		CoinsGranted:      grant.CoinsGranted,
		VkinoCoinsBalance: grant.VKinoCoinsBalance,
	}, nil
}
