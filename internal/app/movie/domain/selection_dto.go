package domain

import (
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
)

type SelectionResponse struct {
	Title  string         `json:"title"`
	Movies []MoviePreview `json:"movies"`
}

func (s *SelectionResponse) Name() string {
	return "selections"
}

func Serialize[T any](v T) ([]byte, error) {
	return serializer.Serialize(v)
}

func Deserialize[T any](data []byte, v *T) error {
	return serializer.Deserialize(data, v)
}
