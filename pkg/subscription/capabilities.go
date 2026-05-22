package subscription

const (
	freeDailyCoinsLimit    int32 = 3
	freeMonthlyRoomLimit   int32 = 3
	level2DailyCoinsLimit  int32 = 6
	level2MonthlyRoomLimit int32 = 6
	level3DailyCoinsLimit  int32 = 12
	level3MonthlyRoomLimit int32 = 10
	level4DailyCoinsLimit  int32 = 30
	defaultMaxRoomMembers  int32 = 2
	extendedMaxRoomMembers int32 = 4
)

type Capabilities struct {
	CanWatchPaidContent bool     `json:"can_watch_paid_content"`
	CanUseSmartContinue bool     `json:"can_use_smart_continue"`
	AdPolicy            AdPolicy `json:"ad_policy"`
	DailyCoinsLimit     int32    `json:"daily_coins_limit"`
	MonthlyRoomLimit    *int32   `json:"monthly_room_limit"`
	MaxRoomMembers      int32    `json:"max_room_members"`
}

type Info struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Level       int32   `json:"level"`
	ActiveUntil *string `json:"active_until"`
}

type Usage struct {
	CoinsReceivedToday      int32  `json:"coins_received_today"`
	CoinsRemainingToday     int32  `json:"coins_remaining_today"`
	RoomsCreatedThisMonth   int32  `json:"rooms_created_this_month"`
	RoomsRemainingThisMonth *int32 `json:"rooms_remaining_this_month"`
}

type State struct {
	Subscription Info         `json:"subscription"`
	Capabilities Capabilities `json:"capabilities"`
	Usage        Usage        `json:"usage"`
}

func FreeCapabilities() Capabilities {
	monthlyRoomLimit := freeMonthlyRoomLimit

	return Capabilities{
		CanWatchPaidContent: false,
		CanUseSmartContinue: false,
		AdPolicy:            AdPolicyNoSkip,
		DailyCoinsLimit:     freeDailyCoinsLimit,
		MonthlyRoomLimit:    &monthlyRoomLimit,
		MaxRoomMembers:      defaultMaxRoomMembers,
	}
}

func Level2Capabilities() Capabilities {
	monthlyRoomLimit := level2MonthlyRoomLimit

	return Capabilities{
		CanWatchPaidContent: true,
		CanUseSmartContinue: true,
		AdPolicy:            AdPolicySkipPreroll,
		DailyCoinsLimit:     level2DailyCoinsLimit,
		MonthlyRoomLimit:    &monthlyRoomLimit,
		MaxRoomMembers:      defaultMaxRoomMembers,
	}
}

func Level3Capabilities() Capabilities {
	monthlyRoomLimit := level3MonthlyRoomLimit

	return Capabilities{
		CanWatchPaidContent: true,
		CanUseSmartContinue: true,
		AdPolicy:            AdPolicySkipAll,
		DailyCoinsLimit:     level3DailyCoinsLimit,
		MonthlyRoomLimit:    &monthlyRoomLimit,
		MaxRoomMembers:      extendedMaxRoomMembers,
	}
}

func Level4Capabilities() Capabilities {
	return Capabilities{
		CanWatchPaidContent: true,
		CanUseSmartContinue: true,
		AdPolicy:            AdPolicyNone,
		DailyCoinsLimit:     level4DailyCoinsLimit,
		MonthlyRoomLimit:    nil,
		MaxRoomMembers:      extendedMaxRoomMembers,
	}
}

func FreeInfo() Info {
	return Info{
		Code:  FreeTariffCode,
		Name:  "Free",
		Level: 1,
	}
}

func DefaultState() State {
	capabilities := FreeCapabilities()

	return State{
		Subscription: FreeInfo(),
		Capabilities: capabilities,
		Usage: Usage{
			CoinsRemainingToday:     capabilities.DailyCoinsLimit,
			RoomsRemainingThisMonth: capabilities.RoomsRemaining(0),
		},
	}
}

func (c Capabilities) CoinsRemaining(receivedToday int32) int32 {
	if receivedToday >= c.DailyCoinsLimit {
		return 0
	}

	return c.DailyCoinsLimit - receivedToday
}

func (c Capabilities) RoomsRemaining(createdThisMonth int32) *int32 {
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
