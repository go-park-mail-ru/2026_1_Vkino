package dto

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type Response struct {
	Email   string `json:"email,omitempty"`
	Message string `json:"message,omitempty"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type AuthContext struct {
	UserID int64
	Email  string
}
