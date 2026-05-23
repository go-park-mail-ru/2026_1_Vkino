package usecase

import (
	"context"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

const (
	freeMonthlyRoomLimit  int32 = 3
	defaultMaxRoomMembers int32 = 2
)

type SubscriptionReader interface {
	GetSubscriptionState(ctx context.Context, userID int64) (*userv1.GetSubscriptionCapabilitiesResponse, error)
}

type userServiceSubscriptionReader struct {
	client userv1.UserServiceClient
}

func NewSubscriptionReader(client userv1.UserServiceClient) SubscriptionReader {
	return &userServiceSubscriptionReader{client: client}
}

func (r *userServiceSubscriptionReader) GetSubscriptionState(
	ctx context.Context,
	userID int64,
) (*userv1.GetSubscriptionCapabilitiesResponse, error) {
	if r == nil || r.client == nil {
		return defaultSubscriptionState(), nil
	}

	if auth, err := authctx.FromContext(ctx); err == nil && auth.Authorization != "" {
		ctx = authctx.AppendOutgoing(ctx, auth.Authorization)
	}

	resp, err := r.client.GetSubscriptionCapabilities(ctx, &userv1.GetSubscriptionCapabilitiesRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return defaultSubscriptionState(), nil
	}

	return resp, nil
}

func defaultSubscriptionState() *userv1.GetSubscriptionCapabilitiesResponse {
	monthlyRoomLimit := freeMonthlyRoomLimit
	roomsRemainingThisMonth := freeMonthlyRoomLimit

	return &userv1.GetSubscriptionCapabilitiesResponse{
		Capabilities: &userv1.SubscriptionCapabilities{
			MonthlyRoomLimit: &monthlyRoomLimit,
			MaxRoomMembers:   defaultMaxRoomMembers,
		},
		Usage: &userv1.SubscriptionUsage{
			RoomsRemainingThisMonth: &roomsRemainingThisMonth,
		},
	}
}
