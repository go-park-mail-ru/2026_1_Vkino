package domain

type SubscriptionTariff struct {
	ID           int64
	Code         string
	Title        string
	Level        int32
	DurationDays int32
}
