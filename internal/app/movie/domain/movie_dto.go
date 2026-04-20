package domain

type MoviePreview struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	ImgUrl string `json:"img_url"`
}

type MovieResponse struct {
	ID                 int64                 `json:"id"`
	Title              string                `json:"title"`
	Description        string                `json:"description"`
	Director           string                `json:"director"`
	TrailerURL         string                `json:"trailer_url"`
	ContentType        string                `json:"content_type"`
	ReleaseYear        int                   `json:"release_year"`
	DurationSeconds    int                   `json:"duration_seconds"`
	AgeLimit           int                   `json:"age_limit"`
	OriginalLanguageID int64                 `json:"original_language_id"`
	OriginalLanguage   string                `json:"original_language"`
	CountryID          int64                 `json:"country_id"`
	Country            string                `json:"country"`
	PictureFileKey     string                `json:"img_url"`
	PosterFileKey      string                `json:"poster_url"`
	Genres             []string              `json:"genres"`
	Actors             []ActorPreview        `json:"actors"`
	Episodes           []EpisodeItemResponse `json:"episodes"`
}
