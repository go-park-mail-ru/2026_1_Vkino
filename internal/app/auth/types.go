package auth

type User struct {
	Email string
	Password string
}

type accessTokenResponse struct {
	AccessToken string
}

type TokenPair struct {
	AccessToken string
	RefreshToken string
}

type SignUpRequest struct {
	Email string
	Password string
}

type SignInRequest struct {
	Email string
	Password string
}
