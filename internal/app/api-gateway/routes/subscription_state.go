package routes

import userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"

const (
	defaultSubscriptionCode        = "free"
	defaultSubscriptionName        = "Free"
	defaultSubscriptionLevel int32 = 1
	defaultDailyCoinsLimit   int32 = 3
	defaultMonthlyRoomLimit  int32 = 3
	defaultMaxRoomMembers    int32 = 2
	defaultAdPolicy                = "no_skip"
)

type subscriptionStateResponse struct {
	Subscription subscriptionInfoResponse         `json:"subscription"`
	Capabilities subscriptionCapabilitiesResponse `json:"capabilities"`
	Usage        subscriptionUsageResponse        `json:"usage"`
}

type subscriptionInfoResponse struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Level       int32   `json:"level"`
	ActiveUntil *string `json:"active_until,omitempty"`
}

type subscriptionCapabilitiesResponse struct {
	CanWatchPaidContent bool   `json:"can_watch_paid_content"`
	CanUseSmartContinue bool   `json:"can_use_smart_continue"`
	AdPolicy            string `json:"ad_policy"`
	DailyCoinsLimit     int32  `json:"daily_coins_limit"`
	MonthlyRoomLimit    *int32 `json:"monthly_room_limit"`
	MaxRoomMembers      int32  `json:"max_room_members"`
}

type subscriptionUsageResponse struct {
	CoinsReceivedToday      int32  `json:"coins_received_today"`
	CoinsRemainingToday     int32  `json:"coins_remaining_today"`
	RoomsCreatedThisMonth   int32  `json:"rooms_created_this_month"`
	RoomsRemainingThisMonth *int32 `json:"rooms_remaining_this_month"`
}

func subscriptionStateFromProto(
	resp *userv1.GetSubscriptionCapabilitiesResponse,
) subscriptionStateResponse {
	if resp == nil {
		return defaultSubscriptionStateResponse()
	}

	state := defaultSubscriptionStateResponse()
	state.Subscription = subscriptionInfoFromProto(resp.GetSubscription())
	state.Capabilities = subscriptionCapabilitiesFromProto(resp.GetCapabilities())
	state.Usage = subscriptionUsageFromProto(resp.GetUsage())

	return state
}

func subscriptionInfoFromProto(info *userv1.SubscriptionInfo) subscriptionInfoResponse {
	defaultInfo := defaultSubscriptionStateResponse().Subscription
	if info == nil {
		return defaultInfo
	}

	defaultInfo.ID = info.GetId()
	defaultInfo.Code = info.GetCode()
	defaultInfo.Name = info.GetName()
	defaultInfo.Level = info.GetLevel()

	if info.ActiveUntil != nil {
		activeUntil := info.GetActiveUntil()
		defaultInfo.ActiveUntil = &activeUntil
	} else {
		defaultInfo.ActiveUntil = nil
	}

	return defaultInfo
}

func subscriptionCapabilitiesFromProto(
	capabilities *userv1.SubscriptionCapabilities,
) subscriptionCapabilitiesResponse {
	defaultCapabilities := defaultSubscriptionStateResponse().Capabilities
	if capabilities == nil {
		return defaultCapabilities
	}

	defaultCapabilities.CanWatchPaidContent = capabilities.GetCanWatchPaidContent()
	defaultCapabilities.CanUseSmartContinue = capabilities.GetCanUseSmartContinue()
	defaultCapabilities.AdPolicy = capabilities.GetAdPolicy()
	defaultCapabilities.DailyCoinsLimit = capabilities.GetDailyCoinsLimit()
	defaultCapabilities.MaxRoomMembers = capabilities.GetMaxRoomMembers()

	if capabilities.MonthlyRoomLimit != nil {
		limit := capabilities.GetMonthlyRoomLimit()
		defaultCapabilities.MonthlyRoomLimit = &limit
	} else {
		defaultCapabilities.MonthlyRoomLimit = nil
	}

	return defaultCapabilities
}

func subscriptionUsageFromProto(usage *userv1.SubscriptionUsage) subscriptionUsageResponse {
	defaultUsage := defaultSubscriptionStateResponse().Usage
	if usage == nil {
		return defaultUsage
	}

	defaultUsage.CoinsReceivedToday = usage.GetCoinsReceivedToday()
	defaultUsage.CoinsRemainingToday = usage.GetCoinsRemainingToday()
	defaultUsage.RoomsCreatedThisMonth = usage.GetRoomsCreatedThisMonth()

	if usage.RoomsRemainingThisMonth != nil {
		remaining := usage.GetRoomsRemainingThisMonth()
		defaultUsage.RoomsRemainingThisMonth = &remaining
	} else {
		defaultUsage.RoomsRemainingThisMonth = nil
	}

	return defaultUsage
}

func defaultSubscriptionStateResponse() subscriptionStateResponse {
	monthlyRoomLimit := defaultMonthlyRoomLimit
	roomsRemainingThisMonth := defaultMonthlyRoomLimit

	return subscriptionStateResponse{
		Subscription: subscriptionInfoResponse{
			Code:  defaultSubscriptionCode,
			Name:  defaultSubscriptionName,
			Level: defaultSubscriptionLevel,
		},
		Capabilities: subscriptionCapabilitiesResponse{
			AdPolicy:         defaultAdPolicy,
			DailyCoinsLimit:  defaultDailyCoinsLimit,
			MonthlyRoomLimit: &monthlyRoomLimit,
			MaxRoomMembers:   defaultMaxRoomMembers,
		},
		Usage: subscriptionUsageResponse{
			CoinsRemainingToday:     defaultDailyCoinsLimit,
			RoomsRemainingThisMonth: &roomsRemainingThisMonth,
		},
	}
}
