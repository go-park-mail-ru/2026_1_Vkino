package domain

type SearchResponse struct {
	Query  string         `json:"query"`
	Movies []MoviePreview `json:"movies"`
	Actors []ActorPreview `json:"actors"`
}
