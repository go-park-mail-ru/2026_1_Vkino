package inmemory

import (
	"errors"
	"log"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
	"github.com/google/uuid"
)

type MovieRepo struct {
	db *DB
}

func NewMovieRepo(db *DB) *MovieRepo {
	s := &MovieRepo{db: db}
	s.initMockData()

	return s
}

var (
	ErrSelectionNotFound = errors.New("selection not found")
	ErrMovieNotFound     = errors.New("movie not found")
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

func (r *MovieRepo) GetMovieByID(id uuid.UUID) (domain.MovieResponse, error) {
	data, err := r.db.Get("movies", id.String())
	if err != nil {
		return domain.MovieResponse{}, ErrMovieNotFound
	}

	var movie domain.MovieResponse
	if err := serializer.Deserialize(data, &movie); err != nil {
		return domain.MovieResponse{}, err
	}

	return movie, nil
}

func (r *MovieRepo) initMockData() {
	movies := []domain.MovieResponse{
		{
			ID:              uuid.New(),
			Title:           "Дюна: Часть Вторая",
			Description:     "Пол Атрейдес, объединившись с народом фрименов и любимой Чани, продолжает свой путь мести тем, кто уничтожил его семью.",
			PictureFileKey:  "img/dune-poster.jpg",
			CoverFileKey:    "img/dune-cover.jpg",
			DurationMinutes: 166,
			AgeLimit:        16,
			ReleaseYear:     2024,
			Country:         "США",
			Director:        "Дени Вильнёв",
			Genres:          []string{"Фантастика", "Драма", "Приключения"},
			Actors: []domain.MovieActorResponse{
				{Name: "Тимати Шаламе", ImgURL: "img/actors/chalamet.jpg"},
				{Name: "Зендея", ImgURL: "img/actors/zendaya.jpg"},
				{Name: "Хавьер Бардем", ImgURL: "img/actors/bardem.jpg"},
				{Name: "Джош Бролин", ImgURL: "img/actors/brolin.jpg"},
			},
		},
		{
			ID:              uuid.New(),
			Title:           "Джокер",
			Description:     "История Артура Флека, который постепенно превращается в Джокера.",
			PictureFileKey:  "img/2.jpeg",
			CoverFileKey:    "img/2.jpeg",
			DurationMinutes: 122,
			AgeLimit:        18,
			ReleaseYear:     2019,
			Country:         "США",
			Director:        "Тодд Филлипс",
			Genres:          []string{"Драма", "Триллер"},
			Actors: []domain.MovieActorResponse{
				{Name: "Хоакин Феникс", ImgURL: "img/actors/phoenix.jpg"},
			},
		},
		{
			ID:              uuid.New(),
			Title:           "Гарри Поттер",
			Description:     "История мальчика-волшебника и его друзей.",
			PictureFileKey:  "img/3.jpg",
			CoverFileKey:    "img/3.jpg",
			DurationMinutes: 152,
			AgeLimit:        12,
			ReleaseYear:     2001,
			Country:         "Великобритания",
			Director:        "Крис Коламбус",
			Genres:          []string{"Фэнтези", "Приключения"},
			Actors: []domain.MovieActorResponse{
				{Name: "Дэниел Рэдклифф", ImgURL: "img/actors/radcliffe.jpg"},
			},
		},
		{
			ID:              uuid.New(),
			Title:           "Little Women",
			Description:     "История взросления четырёх сестёр.",
			PictureFileKey:  "img/4.jpg",
			CoverFileKey:    "img/4.jpg",
			DurationMinutes: 135,
			AgeLimit:        12,
			ReleaseYear:     2019,
			Country:         "США",
			Director:        "Грета Гервиг",
			Genres:          []string{"Драма", "Мелодрама"},
			Actors: []domain.MovieActorResponse{
				{Name: "Сирша Ронан", ImgURL: "img/actors/ronan.jpg"},
			},
		},
		{
			ID:              uuid.New(),
			Title:           "Jaws",
			Description:     "Культовый триллер о гигантской акуле.",
			PictureFileKey:  "img/5.jpg",
			CoverFileKey:    "img/5.jpg",
			DurationMinutes: 124,
			AgeLimit:        16,
			ReleaseYear:     1975,
			Country:         "США",
			Director:        "Стивен Спилберг",
			Genres:          []string{"Триллер", "Приключения"},
			Actors: []domain.MovieActorResponse{
				{Name: "Рой Шайдер", ImgURL: "img/actors/scheider.jpg"},
			},
		},
	}

	for _, movie := range movies {
		movieData, err := serializer.Serialize(movie)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := r.db.Save("movies", movie.ID.String(), movieData); err != nil {
			log.Println(err)
		}
	}

	toPreview := func(m domain.MovieResponse) domain.MoviePreview {
		return domain.MoviePreview{
			ID:             m.ID,
			Title:          m.Title,
			PictureFileKey: m.PictureFileKey,
		}
	}

	selections := map[string]domain.SelectionResponse{
		"popular": {
			Title: "Популярные",
			Movies: []domain.MoviePreview{
				toPreview(movies[0]),
				toPreview(movies[1]),
				toPreview(movies[2]),
				toPreview(movies[3]),
				toPreview(movies[4]),
				toPreview(movies[0]),
				toPreview(movies[1]),
				toPreview(movies[2]),
				toPreview(movies[3]),
				toPreview(movies[4]),
			},
		},
		"new": {
			Title: "Новинки",
			Movies: []domain.MoviePreview{
				toPreview(movies[0]),
				toPreview(movies[1]),
				toPreview(movies[2]),
				toPreview(movies[3]),
				toPreview(movies[4]),
				toPreview(movies[0]),
				toPreview(movies[1]),
				toPreview(movies[2]),
				toPreview(movies[3]),
				toPreview(movies[4]),
			},
		},
		"top": {
			Title: "Топ-10",
			Movies: []domain.MoviePreview{
				toPreview(movies[0]),
				toPreview(movies[1]),
				toPreview(movies[2]),
				toPreview(movies[3]),
				toPreview(movies[4]),
				toPreview(movies[0]),
				toPreview(movies[1]),
				toPreview(movies[2]),
				toPreview(movies[3]),
				toPreview(movies[4]),
			},
		},
	}

	for key, selection := range selections {
		selectionData, err := serializer.Serialize(selection)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := r.db.Save("selections", key, selectionData); err != nil {
			log.Println(err)
		}
	}
}