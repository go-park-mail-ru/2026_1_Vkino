package domain

type SelectionResponse struct {
	Title  string         `json:"title"`
	Movies []MoviePreview `json:"movies"`
}

func (*SelectionResponse) Name() string {
	return "selections"
}
