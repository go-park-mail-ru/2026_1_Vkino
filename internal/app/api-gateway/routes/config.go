package routes

import "time"

type Config interface {
	AuthRequestTimeout() time.Duration
	UserRequestTimeout() time.Duration
	MovieRequestTimeout() time.Duration
	PartyRequestTimeout() time.Duration
	RefreshCookieName() string
	CookieSecure() bool
}
