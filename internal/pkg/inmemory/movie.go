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
	ErrActorNotFound     = errors.New("actor not found")
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

func (r *MovieRepo) GetActorByID(id uuid.UUID) (domain.ActorResponse, error) {
	data, err := r.db.Get("actors", id.String())
	if err != nil {
		return domain.ActorResponse{}, ErrActorNotFound
	}

	var actor domain.ActorResponse
	if err := serializer.Deserialize(data, &actor); err != nil {
		return domain.ActorResponse{}, err
	}

	return actor, nil
}

func (r *MovieRepo) initMockData() {
	// ---- Actors IDs
	chalametID := uuid.New()
	zendayaID := uuid.New()
	bardemID := uuid.New()
	phoenixID := uuid.New()
	baleID := uuid.New()

	// ---- Movies IDs
	duneID := uuid.New()
	jokerID := uuid.New()
	darkKnightID := uuid.New()
	prestigeID := uuid.New()
	fordFerrariID := uuid.New()

	actorPreview := func(id uuid.UUID, name, img string) domain.ActorPreview {
		return domain.ActorPreview{
			ID:             id,
			Name:           name,
			PictureFileKey: img,
		}
	}

	moviePreview := func(id uuid.UUID, title, img string) domain.MoviePreview {
		return domain.MoviePreview{
			ID:             id,
			Title:          title,
			PictureFileKey: img,
		}
	}

	movies := []domain.MovieResponse{
		{
			ID:              duneID,
			Title:           "Дюна: Часть Вторая",
			Description:     "Пол Атрейдес, объединившись с фрименами, продолжает путь мести и принимает судьбоносные решения на Арракисе.",
			PictureFileKey:  "img/1.jpg",
			CoverFileKey:    "img/1.jpg",
			DurationMinutes: 166,
			AgeLimit:        16,
			ReleaseYear:     2024,
			Country:         "США",
			Director:        "Дени Вильнёв",
			Genres:          []string{"Фантастика", "Драма", "Приключения"},
			Actors: []domain.ActorPreview{
				actorPreview(chalametID, "Тимати Шаламе", "img/actors/chalamet.jpg"),
				actorPreview(zendayaID, "Зендея", "img/actors/zendaya.jpg"),
				actorPreview(bardemID, "Хавьер Бардем", "img/actors/bardem.jpg"),
			},
		},
		{
			ID:              jokerID,
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
			Actors: []domain.ActorPreview{
				actorPreview(phoenixID, "Хоакин Феникс", "img/actors/phoenix.jpg"),
			},
		},
		{
			ID:              darkKnightID,
			Title:           "Тёмный рыцарь",
			Description:     "Бэтмен сталкивается с Джокером, который погружает Готэм в хаос.",
			PictureFileKey:  "img/3.jpg",
			CoverFileKey:    "img/3.jpg",
			DurationMinutes: 152,
			AgeLimit:        16,
			ReleaseYear:     2008,
			Country:         "США",
			Director:        "Кристофер Нолан",
			Genres:          []string{"Боевик", "Криминал", "Драма"},
			Actors: []domain.ActorPreview{
				actorPreview(baleID, "Кристиан Бейл", "img/actors/bale.jpg"),
			},
		},
		{
			ID:              prestigeID,
			Title:           "Престиж",
			Description:     "История соперничества двух выдающихся иллюзионистов.",
			PictureFileKey:  "img/4.jpg",
			CoverFileKey:    "img/4.jpg",
			DurationMinutes: 130,
			AgeLimit:        12,
			ReleaseYear:     2006,
			Country:         "США",
			Director:        "Кристофер Нолан",
			Genres:          []string{"Драма", "Триллер", "Детектив"},
			Actors: []domain.ActorPreview{
				actorPreview(baleID, "Кристиан Бейл", "img/actors/bale.jpg"),
			},
		},
		{
			ID:              fordFerrariID,
			Title:           "Ford против Ferrari",
			Description:     "История инженеров и гонщиков, создавших автомобиль, бросивший вызов Ferrari.",
			PictureFileKey:  "img/5.jpg",
			CoverFileKey:    "img/5.jpg",
			DurationMinutes: 152,
			AgeLimit:        12,
			ReleaseYear:     2019,
			Country:         "США",
			Director:        "Джеймс Мэнголд",
			Genres:          []string{"Биография", "Драма", "Спорт"},
			Actors: []domain.ActorPreview{
				actorPreview(baleID, "Кристиан Бейл", "img/actors/bale.jpg"),
			},
		},
	}

	actors := []domain.ActorResponse{
		{
			ID:             chalametID,
			FullName:           "Тимати Шаламе",
			Biography:    "Американский актёр, известный ролями в драматических и фантастических фильмах.",
			BirthDate:      "1995-12-27",
			Country:        "США",
			PictureFileKey: "img/actors/chalamet.jpg",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
			},
		},
		{
			ID:             zendayaID,
			FullName:           "Зендея",
			Biography:    "Американская актриса и певица.",
			BirthDate:      "1996-09-01",
			Country:        "США",
			PictureFileKey: "img/actors/zendaya.jpg",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
			},
		},
		{
			ID:             bardemID,
			FullName:           "Хавьер Бардем",
			Biography:    "Испанский актёр, лауреат премии «Оскар».",
			BirthDate:      "1969-03-01",
			Country:        "Испания",
			PictureFileKey: "img/actors/bardem.jpg",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
			},
		},
		{
			ID:             phoenixID,
			FullName:           "Хоакин Феникс",
			Biography:    "Американский актёр, известный интенсивными драматическими ролями.",
			BirthDate:      "1974-10-28",
			Country:        "США",
			PictureFileKey: "img/actors/phoenix.jpg",
			Movies: []domain.MoviePreview{
				moviePreview(jokerID, "Джокер", "img/2.jpeg"),
			},
		},
		{
			ID:             baleID,
			FullName:       "Кристиан Бейл",
			Biography:      "Британский актёр, известный ролями в психологически и физически сложных образах.",
			BirthDate:      "1974-01-30",
			Country:        "Великобритания",
			PictureFileKey: "img/actors/bale.jpg",
			Movies: []domain.MoviePreview{
				moviePreview(darkKnightID, "Тёмный рыцарь", "img/3.jpg"),
				moviePreview(prestigeID, "Престиж", "img/4.jpg"),
				moviePreview(fordFerrariID, "Ford против Ferrari", "img/5.jpg"),
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

	for _, actor := range actors {
		actorData, err := serializer.Serialize(actor)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := r.db.Save("actors", actor.ID.String(), actorData); err != nil {
			log.Println(err)
		}
	}

	selections := map[string]domain.SelectionResponse{
		"popular": {
			Title: "Популярные",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
				moviePreview(jokerID, "Джокер", "img/2.jpeg"),
				moviePreview(darkKnightID, "Тёмный рыцарь", "img/3.jpg"),
				moviePreview(prestigeID, "Престиж", "img/4.jpg"),
				moviePreview(fordFerrariID, "Ford против Ferrari", "img/5.jpg"),
			},
		},
		"new": {
			Title: "Новинки",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
				moviePreview(jokerID, "Джокер", "img/2.jpeg"),
				moviePreview(darkKnightID, "Тёмный рыцарь", "img/3.jpg"),
				moviePreview(prestigeID, "Престиж", "img/4.jpg"),
				moviePreview(fordFerrariID, "Ford против Ferrari", "img/5.jpg"),
			},
		},
		"top": {
			Title: "Топ-10",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
				moviePreview(jokerID, "Джокер", "img/2.jpeg"),
				moviePreview(darkKnightID, "Тёмный рыцарь", "img/3.jpg"),
				moviePreview(prestigeID, "Престиж", "img/4.jpg"),
				moviePreview(fordFerrariID, "Ford против Ferrari", "img/5.jpg"),
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