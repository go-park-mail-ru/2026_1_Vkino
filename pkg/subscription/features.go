package subscription

type FeatureCode string

const (
	FeaturePaidContent     FeatureCode = "paid_content"
	FeatureSmartContinue   FeatureCode = "smart_continue"
	FeatureAdSkip          FeatureCode = "ad_skip"
	FeatureWatchPartyRooms FeatureCode = "watch_party_rooms"
	FeatureWatchPartyUsers FeatureCode = "watch_party_members"
	FeatureDailyCoins      FeatureCode = "daily_coins"
)

const (
	OptionCanWatchPaidContent = "can_watch_paid_content"
	OptionCanUseSmartContinue = "can_use_smart_continue"
	OptionAdPolicy            = "ad_policy"
	OptionDailyCoinsLimit     = "daily_coins_limit"
	OptionMonthlyRoomLimit    = "monthly_room_limit"
	OptionMaxRoomMembers      = "max_room_members"
)

const FreeTariffCode = "free"
