package domain

type FavoritesResponse struct {
	MovieIDs   []int64 `json:"movie_ids"`
	TotalCount int32   `json:"total_count"`
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
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UserSearchResult struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
	IsFriend  bool   `json:"is_friend"`
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
