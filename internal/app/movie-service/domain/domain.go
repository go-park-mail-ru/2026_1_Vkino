package domain

import (
	"time"
)

type Movie struct {
	ID                 int64
	Title              string
	Description        string
	Director           string
	TrailerURL         string
	ContentType        string
	ReleaseYear        int
	DurationSeconds    int
	AgeLimit           int
	OriginalLanguageID int64
	OriginalLanguage   string
	CountryID          int64
	Country            string
	PictureFileKey     string
	PosterFileKey      string
	Genres             []string
	Actors             []ActorShort
	Episodes           []Episode
}

type Actor struct {
	ID             int64
	FullName       string
	BirthDate      *time.Time
	Biography      string
	CountryID      int64
	PictureFileKey string
	Movies         []MovieCard
}

type ActorShort struct {
	ID             int64
	FullName       string
	PictureFileKey string
}

type Episode struct {
	ID              int64
	MovieID         int64
	SeasonNumber    int
	EpisodeNumber   int
	Title           string
	Description     string
	DurationSeconds int
	PictureFileKey  string
	VideoFileKey    string
}

type MovieCard struct {
	ID             int64
	Title          string
	PictureFileKey string
}

type Genre struct {
	ID     int64
	Title  string
	Movies []MovieCard
}

type Selection struct {
	Title  string
	Movies []MovieCard
}

type EpisodeProgress struct {
	EpisodeID       int64
	PositionSeconds int64
}

type WatchProgressItem struct {
	EpisodeID       int64
	MovieID         int64
	MovieTitle      string
	PosterURL       string
	ContentType     string
	SeasonNumber    int
	EpisodeNumber   int
	EpisodeTitle    string
	PositionSeconds int64
	DurationSeconds int64
	UpdatedAt       string
}
