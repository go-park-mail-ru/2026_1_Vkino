package domain

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthContext struct {
	UserID int64
	Email  string
}
