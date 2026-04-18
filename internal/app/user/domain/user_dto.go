package domain

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
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

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type UpdateProfileRequest struct {
	Birthdate *string `json:"birthdate"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type FavoriteMovieResponse struct {
	MovieID    int64 `json:"movie_id"`
	IsFavorite bool  `json:"is_favorite"`
}

type ProfileResponse struct {
	Email     string  `json:"email"`
	Birthdate *string `json:"birthdate,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
}

func (u *TokenPair) Name() string {
	return "sessions"
}
