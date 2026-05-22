package subscription

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Option struct {
	Code  string
	Value *string
}

type optionApplier func(*Capabilities, Option) error

var ErrUnsupportedAdPolicy = errors.New("unsupported ad policy")

var optionAppliers = map[string]optionApplier{
	OptionCanWatchPaidContent: applyBoolOption(func(capabilities *Capabilities, value bool) {
		capabilities.CanWatchPaidContent = value
	}),
	OptionCanUseSmartContinue: applyBoolOption(func(capabilities *Capabilities, value bool) {
		capabilities.CanUseSmartContinue = value
	}),
	OptionAdPolicy: applyAdPolicyOption,
	OptionDailyCoinsLimit: applyInt32Option(func(capabilities *Capabilities, value int32) {
		capabilities.DailyCoinsLimit = value
	}),
	OptionMonthlyRoomLimit: applyMonthlyRoomLimitOption,
	OptionMaxRoomMembers: applyInt32Option(func(capabilities *Capabilities, value int32) {
		capabilities.MaxRoomMembers = value
	}),
}

func ParseCapabilities(options []Option) (Capabilities, error) {
	capabilities := FreeCapabilities()

	for _, option := range options {
		code := strings.TrimSpace(option.Code)

		apply, ok := optionAppliers[code]
		if !ok {
			continue
		}

		if err := apply(&capabilities, option); err != nil {
			return Capabilities{}, fmt.Errorf("parse subscription option %q: %w", option.Code, err)
		}
	}

	return capabilities, nil
}

func applyBoolOption(setter func(*Capabilities, bool)) optionApplier {
	return func(capabilities *Capabilities, option Option) error {
		value, err := parseBoolOption(option)
		if err != nil {
			return err
		}

		setter(capabilities, value)

		return nil
	}
}

func applyInt32Option(setter func(*Capabilities, int32)) optionApplier {
	return func(capabilities *Capabilities, option Option) error {
		value, err := parseInt32Option(option)
		if err != nil {
			return err
		}

		setter(capabilities, value)

		return nil
	}
}

func applyAdPolicyOption(capabilities *Capabilities, option Option) error {
	value, err := parseAdPolicyOption(option)
	if err != nil {
		return err
	}

	capabilities.AdPolicy = value

	return nil
}

func applyMonthlyRoomLimitOption(capabilities *Capabilities, option Option) error {
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

func parseBoolOption(option Option) (bool, error) {
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

func parseInt32Option(option Option) (int32, error) {
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

func parseNullableInt32Option(option Option) (int32, bool, error) {
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

func parseAdPolicyOption(option Option) (AdPolicy, error) {
	if option.Value == nil {
		return AdPolicyNoSkip, nil
	}

	value := AdPolicy(strings.TrimSpace(*option.Value))

	switch value {
	case AdPolicyNoSkip, AdPolicySkipPreroll, AdPolicySkipAll, AdPolicyNone:
		return value, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedAdPolicy, value)
	}
}
