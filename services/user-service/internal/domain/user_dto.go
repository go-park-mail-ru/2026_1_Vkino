package domain

type FavoriteMovieResponse struct {
	MovieID    int64 `json:"movie_id"`
	IsFavorite bool  `json:"is_favorite"`
}

type ProfileResponse struct {
	Email     string  `json:"email"`
	Birthdate *string `json:"birthdate"`
	AvatarURL string  `json:"avatar_url"`
}

type FriendResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type UserSearchResult struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	IsFriend bool   `json:"is_friend"`
}