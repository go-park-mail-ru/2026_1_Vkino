package domain

type MovieResponse struct {
	ID                 int64                `json:"id"`
	Title              string               `json:"title"`
	Description        string               `json:"description"`
	Director           string               `json:"director"`
	TrailerURL         string               `json:"trailer_url"`
	ContentType        string               `json:"content_type"`
	ReleaseYear        int                  `json:"release_year"`
	DurationSeconds    int                  `json:"duration_seconds"`
	AgeLimit           int                  `json:"age_limit"`
	OriginalLanguageID int64                `json:"original_language_id"`
	OriginalLanguage   string               `json:"original_language"`
	CountryID          int64                `json:"country_id"`
	Country            string               `json:"country"`
	PictureFileKey     string               `json:"img_url"`
	PosterFileKey      string               `json:"poster_url"`
	Genres             []string             `json:"genres"`
	Actors             []ActorShortResponse `json:"actors"`
	Episodes           []EpisodeResponse    `json:"episodes"`
	IsFavorite         bool                 `json:"is_favorite"`
}

type ActorResponse struct {
	ID             int64               `json:"id"`
	FullName       string              `json:"full_name"`
	Biography      string              `json:"biography"`
	BirthDate      string              `json:"birthdate"`
	Country        string              `json:"country"`
	PictureFileKey string              `json:"img_url"`
	Movies         []MovieCardResponse `json:"movies"`
}

type ActorShortResponse struct {
	ID             int64  `json:"id"`
	FullName       string `json:"full_name"`
	PictureFileKey string `json:"img_url"`
}

type EpisodeResponse struct {
	ID              int64  `json:"id"`
	MovieID         int64  `json:"movie_id"`
	SeasonNumber    int    `json:"season_number"`
	EpisodeNumber   int    `json:"episode_number"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	DurationSeconds int    `json:"duration_seconds"`
	PictureFileKey  string `json:"img_url"`
	VideoURL        string `json:"video_url"`
}

type MovieCardResponse struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	PictureFileKey string `json:"img_url"`
}

type SearchResponse struct {
	Movies []MovieCardResponse  `json:"movies"`
	Actors []ActorShortResponse `json:"actors"`
}

type GenreResponse struct {
	ID     int64               `json:"id"`
	Title  string              `json:"title"`
	Movies []MovieCardResponse `json:"movies"`
}

type GenreShortResponse struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type SelectionResponse struct {
	Title  string              `json:"title"`
	Movies []MovieCardResponse `json:"movies"`
}

type EpisodePlaybackResponse struct {
	EpisodeID       int64  `json:"episode_id"`
	MovieID         int64  `json:"movie_id"`
	SeasonNumber    int    `json:"season_number"`
	EpisodeNumber   int    `json:"episode_number"`
	Title           string `json:"title"`
	DurationSeconds int    `json:"duration_seconds"`
	PlaybackURL     string `json:"playback_url"`
	PositionSeconds int64  `json:"position_seconds"`
}

type EpisodeProgressResponse struct {
	EpisodeID       int64 `json:"episode_id"`
	PositionSeconds int64 `json:"position_seconds"`
}

type WatchProgressItemResponse struct {
	EpisodeID       int64  `json:"episode_id"`
	MovieID         int64  `json:"movie_id"`
	MovieTitle      string `json:"movie_title"`
	PosterURL       string `json:"poster_url"`
	ContentType     string `json:"content_type"`
	SeasonNumber    int    `json:"season_number"`
	EpisodeNumber   int    `json:"episode_number"`
	EpisodeTitle    string `json:"episode_title"`
	PositionSeconds int64  `json:"position_seconds"`
	DurationSeconds int64  `json:"duration_seconds"`
	UpdatedAt       string `json:"updated_at"`
}
