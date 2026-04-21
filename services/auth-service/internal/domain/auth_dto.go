package domain

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthContext struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}