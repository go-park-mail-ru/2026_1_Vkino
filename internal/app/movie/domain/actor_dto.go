package domain

type ActorPreview struct {
	ID             int64  `json:"id"`
	FullName       string `json:"full_name"`
	PictureFileKey string `json:"img_url"`
}

type ActorResponse struct {
	ID             int64          `json:"id"`
	FullName       string         `json:"full_name"`
	Biography      string         `json:"biography"`
	BirthDate      string         `json:"birth_date"`
	CountryID      int64          `json:"country_id"`
	PictureFileKey string         `json:"img_url"`
	Movies         []MoviePreview `json:"movies"`
}

func (a *ActorResponse) Name() string {
	return "actors"
}
