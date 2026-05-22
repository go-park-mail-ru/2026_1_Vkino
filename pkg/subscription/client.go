package subscription

import (
	"context"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

type StateReader interface {
	GetSubscriptionState(ctx context.Context, userID int64) (State, error)
}

type userServiceStateReader struct {
	client userv1.UserServiceClient
}

func NewStateReader(client userv1.UserServiceClient) StateReader {
	return &userServiceStateReader{client: client}
}

func (r *userServiceStateReader) GetSubscriptionState(ctx context.Context, userID int64) (State, error) {
	if r == nil || r.client == nil {
		return DefaultState(), nil
	}

	if auth, err := authctx.FromContext(ctx); err == nil && auth.Authorization != "" {
		ctx = authctx.AppendOutgoing(ctx, auth.Authorization)
	}

	resp, err := r.client.GetSubscriptionCapabilities(ctx, &userv1.GetSubscriptionCapabilitiesRequest{
		UserId: userID,
	})
	if err != nil {
		return State{}, err
	}

	return StateFromProto(resp), nil
}

func StateFromProto(resp *userv1.GetSubscriptionCapabilitiesResponse) State {
	if resp == nil {
		return DefaultState()
	}

	state := DefaultState()
	applySubscriptionInfoFromProto(&state, resp.GetSubscription())
	applyCapabilitiesFromProto(&state, resp.GetCapabilities())
	applyUsageFromProto(&state, resp.GetUsage())

	return state
}

func applySubscriptionInfoFromProto(state *State, subscriptionInfo *userv1.SubscriptionInfo) {
	if subscriptionInfo == nil {
		return
	}

	state.Subscription = Info{
		ID:    subscriptionInfo.GetId(),
		Code:  subscriptionInfo.GetCode(),
		Name:  subscriptionInfo.GetName(),
		Level: subscriptionInfo.GetLevel(),
	}

	if subscriptionInfo.ActiveUntil != nil {
		activeUntil := subscriptionInfo.GetActiveUntil()
		state.Subscription.ActiveUntil = &activeUntil
	}
}

func applyCapabilitiesFromProto(state *State, capabilities *userv1.SubscriptionCapabilities) {
	if capabilities == nil {
		return
	}

	state.Capabilities = Capabilities{
		CanWatchPaidContent: capabilities.GetCanWatchPaidContent(),
		CanUseSmartContinue: capabilities.GetCanUseSmartContinue(),
		AdPolicy:            AdPolicy(capabilities.GetAdPolicy()),
		DailyCoinsLimit:     capabilities.GetDailyCoinsLimit(),
		MaxRoomMembers:      capabilities.GetMaxRoomMembers(),
	}

	if capabilities.MonthlyRoomLimit != nil {
		limit := capabilities.GetMonthlyRoomLimit()
		state.Capabilities.MonthlyRoomLimit = &limit
	} else {
		state.Capabilities.MonthlyRoomLimit = nil
	}
}

func applyUsageFromProto(state *State, usage *userv1.SubscriptionUsage) {
	if usage == nil {
		return
	}

	state.Usage = Usage{
		CoinsReceivedToday:    usage.GetCoinsReceivedToday(),
		CoinsRemainingToday:   usage.GetCoinsRemainingToday(),
		RoomsCreatedThisMonth: usage.GetRoomsCreatedThisMonth(),
	}

	if usage.RoomsRemainingThisMonth != nil {
		remaining := usage.GetRoomsRemainingThisMonth()
		state.Usage.RoomsRemainingThisMonth = &remaining
	} else {
		state.Usage.RoomsRemainingThisMonth = nil
	}
}
