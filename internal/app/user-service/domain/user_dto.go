package domain

type FavoritesResponse struct {
	Movies     []MovieCardResponse `json:"movies"`
	TotalCount int32               `json:"total_count"`
}

type MovieCardResponse struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	PictureFileKey string `json:"img_url"`
}

type FavoriteMovieResponse struct {
	MovieID    int64 `json:"movie_id"`
	IsFavorite bool  `json:"is_favorite"`
}

type ProfileResponse struct {
	Email     string  `json:"email"`
	Role      string  `json:"role"`
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

type FriendRequestItem struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type FriendsListResponse struct {
	Friends    []UserSearchResult `json:"friends"`
	TotalCount int32              `json:"total_count"`
}
