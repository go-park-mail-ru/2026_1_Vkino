package clock

import "time"

type Service interface {
	Now() time.Time
}

type RealClock struct{}

func New() *RealClock {
	return &RealClock{}
}

func (c *RealClock) Now() time.Time {
	return time.Now()
}