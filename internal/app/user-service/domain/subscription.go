package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	SubscriptionFreeTariffCode = "free"

	SubscriptionOptionCanWatchPaidContent = "can_watch_paid_content"
	SubscriptionOptionCanUseSmartContinue = "can_use_smart_continue"
	SubscriptionOptionAdPolicy            = "ad_policy"
	SubscriptionOptionDailyCoinsLimit     = "daily_coins_limit"
	SubscriptionOptionMonthlyRoomLimit    = "monthly_room_limit"
	SubscriptionOptionMaxRoomMembers      = "max_room_members"

	defaultDailyCoinsLimit  int32 = 3
	defaultMonthlyRoomLimit int32 = 3
	defaultMaxRoomMembers   int32 = 2
)

type SubscriptionAdPolicy string

const (
	SubscriptionAdPolicyNoSkip      SubscriptionAdPolicy = "no_skip"
	SubscriptionAdPolicySkipPreroll SubscriptionAdPolicy = "skip_preroll"
	SubscriptionAdPolicySkipAll     SubscriptionAdPolicy = "skip_all"
	SubscriptionAdPolicyNone        SubscriptionAdPolicy = "none"
)

type SubscriptionOption struct {
	Code  string
	Value *string
}

type SubscriptionInfo struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Level       int32   `json:"level"`
	ActiveUntil *string `json:"active_until"`
}

type SubscriptionCapabilities struct {
	CanWatchPaidContent bool                 `json:"can_watch_paid_content"`
	CanUseSmartContinue bool                 `json:"can_use_smart_continue"`
	AdPolicy            SubscriptionAdPolicy `json:"ad_policy"`
	DailyCoinsLimit     int32                `json:"daily_coins_limit"`
	MonthlyRoomLimit    *int32               `json:"monthly_room_limit"`
	MaxRoomMembers      int32                `json:"max_room_members"`
}

type SubscriptionUsage struct {
	CoinsReceivedToday      int32  `json:"coins_received_today"`
	CoinsRemainingToday     int32  `json:"coins_remaining_today"`
	RoomsCreatedThisMonth   int32  `json:"rooms_created_this_month"`
	RoomsRemainingThisMonth *int32 `json:"rooms_remaining_this_month"`
}

type SubscriptionState struct {
	Subscription SubscriptionInfo         `json:"subscription"`
	Capabilities SubscriptionCapabilities `json:"capabilities"`
	Usage        SubscriptionUsage        `json:"usage"`
}

type subscriptionOptionApplier func(*SubscriptionCapabilities, SubscriptionOption) error

var ErrUnsupportedSubscriptionAdPolicy = errors.New("unsupported subscription ad policy")

var subscriptionOptionAppliers = map[string]subscriptionOptionApplier{
	SubscriptionOptionCanWatchPaidContent: applyBoolOption(func(capabilities *SubscriptionCapabilities, value bool) {
		capabilities.CanWatchPaidContent = value
	}),
	SubscriptionOptionCanUseSmartContinue: applyBoolOption(func(capabilities *SubscriptionCapabilities, value bool) {
		capabilities.CanUseSmartContinue = value
	}),
	SubscriptionOptionAdPolicy: applyAdPolicyOption,
	SubscriptionOptionDailyCoinsLimit: applyInt32Option(func(capabilities *SubscriptionCapabilities, value int32) {
		capabilities.DailyCoinsLimit = value
	}),
	SubscriptionOptionMonthlyRoomLimit: applyMonthlyRoomLimitOption,
	SubscriptionOptionMaxRoomMembers: applyInt32Option(func(capabilities *SubscriptionCapabilities, value int32) {
		capabilities.MaxRoomMembers = value
	}),
}

func DefaultSubscriptionState() SubscriptionState {
	monthlyRoomLimit := defaultMonthlyRoomLimit
	capabilities := SubscriptionCapabilities{
		AdPolicy:         SubscriptionAdPolicyNoSkip,
		DailyCoinsLimit:  defaultDailyCoinsLimit,
		MonthlyRoomLimit: &monthlyRoomLimit,
		MaxRoomMembers:   defaultMaxRoomMembers,
	}

	return SubscriptionState{
		Subscription: SubscriptionInfo{
			Code:  SubscriptionFreeTariffCode,
			Name:  "Free",
			Level: 1,
		},
		Capabilities: capabilities,
		Usage: SubscriptionUsage{
			CoinsRemainingToday:     capabilities.CoinsRemaining(0),
			RoomsRemainingThisMonth: capabilities.RoomsRemaining(0),
		},
	}
}

func ParseSubscriptionCapabilities(options []SubscriptionOption) (SubscriptionCapabilities, error) {
	capabilities := SubscriptionCapabilities{
		AdPolicy: SubscriptionAdPolicyNoSkip,
	}

	for _, option := range options {
		code := strings.TrimSpace(option.Code)

		apply, ok := subscriptionOptionAppliers[code]
		if !ok {
			continue
		}

		if err := apply(&capabilities, option); err != nil {
			return SubscriptionCapabilities{}, fmt.Errorf(
				"parse subscription option %q: %w",
				option.Code,
				err,
			)
		}
	}

	return capabilities, nil
}

func (c SubscriptionCapabilities) CoinsRemaining(receivedToday int32) int32 {
	if receivedToday >= c.DailyCoinsLimit {
		return 0
	}

	return c.DailyCoinsLimit - receivedToday
}

func (c SubscriptionCapabilities) RoomsRemaining(createdThisMonth int32) *int32 {
	if c.MonthlyRoomLimit == nil {
		return nil
	}

	if createdThisMonth >= *c.MonthlyRoomLimit {
		zero := int32(0)

		return &zero
	}

	remaining := *c.MonthlyRoomLimit - createdThisMonth

	return &remaining
}

func applyBoolOption(setter func(*SubscriptionCapabilities, bool)) subscriptionOptionApplier {
	return func(capabilities *SubscriptionCapabilities, option SubscriptionOption) error {
		value, err := parseBoolOption(option)
		if err != nil {
			return err
		}

		setter(capabilities, value)

		return nil
	}
}

func applyInt32Option(setter func(*SubscriptionCapabilities, int32)) subscriptionOptionApplier {
	return func(capabilities *SubscriptionCapabilities, option SubscriptionOption) error {
		value, err := parseInt32Option(option)
		if err != nil {
			return err
		}

		setter(capabilities, value)

		return nil
	}
}

func applyAdPolicyOption(capabilities *SubscriptionCapabilities, option SubscriptionOption) error {
	value, err := parseAdPolicyOption(option)
	if err != nil {
		return err
	}

	capabilities.AdPolicy = value

	return nil
}

func applyMonthlyRoomLimitOption(capabilities *SubscriptionCapabilities, option SubscriptionOption) error {
	value, hasValue, err := parseNullableInt32Option(option)
	if err != nil {
		return err
	}

	if hasValue {
		capabilities.MonthlyRoomLimit = &value
	} else {
		capabilities.MonthlyRoomLimit = nil
	}

	return nil
}

func parseBoolOption(option SubscriptionOption) (bool, error) {
	if option.Value == nil {
		return false, nil
	}

	value := strings.TrimSpace(*option.Value)
	if value == "" {
		return false, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}

	return parsed, nil
}

func parseInt32Option(option SubscriptionOption) (int32, error) {
	if option.Value == nil {
		return 0, nil
	}

	value := strings.TrimSpace(*option.Value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(parsed), nil
}

func parseNullableInt32Option(option SubscriptionOption) (int32, bool, error) {
	if option.Value == nil {
		return 0, false, nil
	}

	value := strings.TrimSpace(*option.Value)
	if value == "" || strings.EqualFold(value, "null") {
		return 0, false, nil
	}

	parsed, err := parseInt32Option(option)
	if err != nil {
		return 0, false, err
	}

	return parsed, true, nil
}

func parseAdPolicyOption(option SubscriptionOption) (SubscriptionAdPolicy, error) {
	if option.Value == nil {
		return SubscriptionAdPolicyNoSkip, nil
	}

	value := SubscriptionAdPolicy(strings.TrimSpace(*option.Value))

	switch value {
	case SubscriptionAdPolicyNoSkip,
		SubscriptionAdPolicySkipPreroll,
		SubscriptionAdPolicySkipAll,
		SubscriptionAdPolicyNone:
		return value, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedSubscriptionAdPolicy, value)
	}
}
