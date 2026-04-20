package domain

type FavoriteMovieResponse struct {
	MovieID    int64
	IsFavorite bool
}

type ProfileResponse struct {
	Email     string
	Birthdate *string
	AvatarURL string
}

type FriendResponse struct {
	ID    int64
	Email string
}

type UserSearchResult struct {
	ID       int64
	Email    string
	IsFriend bool
}
