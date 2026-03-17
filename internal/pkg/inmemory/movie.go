package inmemory

import (
	"errors"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
	"github.com/google/uuid"
)

type MovieRepo struct {
	db *DB
}

func NewMovieRepo(db *DB) *MovieRepo {
	s := &MovieRepo{db: db}
	s.initMockSelections()

	return s
}

var (
	ErrSelectionNotFound = errors.New("selection not found")
)

func (r *MovieRepo) GetSelectionByTitle(title string) (domain.SelectionResponse, error) {
	data, err := r.db.Get("selections", title)
	if err != nil {
		return domain.SelectionResponse{}, ErrSelectionNotFound
	}

	var selection domain.SelectionResponse
	if err := serializer.Deserialize(data, &selection); err != nil {
		return domain.SelectionResponse{}, err
	}

	return selection, nil
}

func (r *MovieRepo) GetAllSelections() ([]domain.SelectionResponse, error) {
	allData, err := r.db.GetAll("selections")
	if err != nil {
		return []domain.SelectionResponse{}, err
	}

	var selections []domain.SelectionResponse

	for _, data := range allData {
		var sel domain.SelectionResponse
		if err := serializer.Deserialize(data, &sel); err == nil {
			selections = append(selections, sel)
		}
	}

	return selections, nil
}

func (r *MovieRepo) initMockSelections() {
	movies := []domain.MoviePreview{
		{
			ID:             uuid.New(),
			Title:          "65",
			PictureFileKey: "img/1.jpg",
		},
		{
			ID:             uuid.New(),
			Title:          "Джокер",
			PictureFileKey: "img/2.jpeg",
		},
		{
			ID:             uuid.New(),
			Title:          "Гарри Поттер",
			PictureFileKey: "img/3.jpg",
		},
		{
			ID:             uuid.New(),
			Title:          "Little Women",
			PictureFileKey: "img/4.jpg",
		},
		{
			ID:             uuid.New(),
			Title:          "Jaws",
			PictureFileKey: "img/5.jpg",
		},
		{
			ID:             uuid.New(),
			Title:          "65",
			PictureFileKey: "img/1.jpg",
		},
		{
			ID:             uuid.New(),
			Title:          "Джокер",
			PictureFileKey: "img/2.jpeg",
		},
		{
			ID:             uuid.New(),
			Title:          "Гарри Поттер",
			PictureFileKey: "img/3.jpg",
		},
		{
			ID:             uuid.New(),
			Title:          "Little Women",
			PictureFileKey: "img/4.jpg",
		},
		{
			ID:             uuid.New(),
			Title:          "Jaws",
			PictureFileKey: "img/5.jpg",
		},
	}

	selections := map[string]domain.SelectionResponse{
		"popular": {
			Title: "Популярные",
			Movies: []domain.MoviePreview{
				movies[0], movies[1], movies[2], movies[3], movies[4], movies[0], movies[1], movies[2], movies[3], movies[4],
			},
		},
		"new": {
			Title: "Новинки",
			Movies: []domain.MoviePreview{
				movies[0], movies[1], movies[2], movies[3], movies[4], movies[0], movies[1], movies[2], movies[3], movies[4],
			},
		},
		"top": {
			Title: "Топ-10",
			Movies: []domain.MoviePreview{
				movies[0], movies[1], movies[2], movies[3], movies[4], movies[0], movies[1], movies[2], movies[3], movies[4],
			},
		},
	}

	for key, selection := range selections {
		selectionData, _ := serializer.Serialize(selection)
		r.db.Save("selections", key, selectionData)
	}
}
