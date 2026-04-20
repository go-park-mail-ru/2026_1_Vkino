package dto

type ProfileResponse struct {
	Email     string  `json:"email"`
	Birthdate *string `json:"birthdate,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
}

type UpdateProfileRequest struct {
	Birthdate string `json:"birthdate"`
}

type SearchUsersResponse struct {
	Users []UserSearchResult `json:"users"`
}

type UserSearchResult struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	IsFriend bool   `json:"is_friend"`
}

type AddFriendRequest struct {
	FriendID int64 `json:"friend_id"`
}

type FriendResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type DeleteFriendRequest struct {
	FriendID int64 `json:"friend_id"`
}

type DeleteFriendResponse struct {
	Success bool `json:"success"`
}

type AddMovieToFavoritesRequest struct {
	MovieID int64 `json:"movie_id"`
}

type FavoriteMovieResponse struct {
	MovieID    int64 `json:"movie_id"`
	IsFavorite bool  `json:"is_favorite"`
}
