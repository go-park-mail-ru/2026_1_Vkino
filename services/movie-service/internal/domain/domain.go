package domain

import "strings"

type Movie struct {
	ID            int64
	Title         string
	Description   string
	Year          int
	AgeLimit      int
	DurationMin   int
	PosterFileKey *string
	CardFileKey   *string
	Countries     []string
	Genres        []string
	Actors        []ActorShort
	Episodes      []Episode
}

type Actor struct {
	ID            int64
	Name          string
	Description   string
	AvatarFileKey *string
	Movies        []MovieCard
}

type ActorShort struct {
	ID            int64
	Name          string
	AvatarFileKey *string
}

type Episode struct {
	ID          int64
	Number      int
	Title       string
	DurationSec int
}

type MovieCard struct {
	ID            int64
	Title         string
	Year          int
	PosterFileKey *string
	CardFileKey   *string
}

type Selection struct {
	Title  string
	Movies []MovieCard
}

func ValidateSelectionTitle(title string) bool {
	trimmed := strings.TrimSpace(title)
	return trimmed != "" && len(trimmed) <= 255
}
