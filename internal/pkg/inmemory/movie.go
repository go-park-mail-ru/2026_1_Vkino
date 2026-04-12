package inmemory

import (
	"errors"
	"log"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
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

func (r *MovieRepo) GetMovieByID(id int64) (domain.MovieResponse, error) {
	data, err := r.db.Get("movies", strconv.FormatInt(id, 10))
	if err != nil {
		return domain.MovieResponse{}, ErrMovieNotFound
	}

	var movie domain.MovieResponse
	if err := serializer.Deserialize(data, &movie); err != nil {
		return domain.MovieResponse{}, err
	}

	return movie, nil
}

func (r *MovieRepo) GetActorByID(id int64) (domain.ActorResponse, error) {
	data, err := r.db.Get("actors", strconv.FormatInt(id, 10))
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
	const (
		countryUSA      = 1
		countrySpain    = 2
		countryUK       = 3
		languageEnglish = 1
		contentTypeFilm = "film"
	)

	var (
		chalametID    int64 = 1
		zendayaID     int64 = 2
		bardemID      int64 = 3
		phoenixID     int64 = 4
		baleID        int64 = 5
		duneID        int64 = 101
		jokerID       int64 = 102
		darkKnightID  int64 = 103
		prestigeID    int64 = 104
		fordFerrariID int64 = 105
	)

	actorPreview := func(id int64, fullName, img string) domain.ActorPreview {
		return domain.ActorPreview{
			ID:             id,
			FullName:       fullName,
			PictureFileKey: img,
		}
	}

	moviePreview := func(id int64, title, img string) domain.MoviePreview {
		return domain.MoviePreview{
			ID:     id,
			Title:  title,
			ImgUrl: img,
		}
	}

	movies := []domain.MovieResponse{
		{
			ID:    duneID,
			Title: "Дюна: Часть Вторая",
			Description: "Пол Атрейдес, объединившись с фрименами, продолжает путь мести и " +
				"принимает судьбоносные решения на Арракисе.",
			Director:           "Дени Вильнёв",
			ContentType:        contentTypeFilm,
			ReleaseYear:        2024,
			DurationSeconds:    166 * 60,
			AgeLimit:           16,
			OriginalLanguageID: languageEnglish,
			CountryID:          countryUSA,
			PictureFileKey:     "img/1.jpg",
			Genres:             []string{"Фантастика", "Драма", "Приключения"},
			Actors: []domain.ActorPreview{
				actorPreview(chalametID, "Тимати Шаламе", "img/actors/chalamet.jpg"),
				actorPreview(zendayaID, "Зендея", "img/actors/zendaya.jpg"),
				actorPreview(bardemID, "Хавьер Бардем", "img/actors/bardem.jpg"),
			},
		},
		{
			ID:                 jokerID,
			Title:              "Джокер",
			Description:        "История Артура Флека, который постепенно превращается в Джокера.",
			Director:           "Тодд Филлипс",
			ContentType:        contentTypeFilm,
			ReleaseYear:        2019,
			DurationSeconds:    122 * 60,
			AgeLimit:           18,
			OriginalLanguageID: languageEnglish,
			CountryID:          countryUSA,
			PictureFileKey:     "img/2.jpeg",
			Genres:             []string{"Драма", "Триллер"},
			Actors: []domain.ActorPreview{
				actorPreview(phoenixID, "Хоакин Феникс", "img/actors/phoenix.jpg"),
			},
		},
		{
			ID:                 darkKnightID,
			Title:              "Тёмный рыцарь",
			Description:        "Бэтмен сталкивается с Джокером, который погружает Готэм в хаос.",
			Director:           "Кристофер Нолан",
			ContentType:        contentTypeFilm,
			ReleaseYear:        2008,
			DurationSeconds:    152 * 60,
			AgeLimit:           16,
			OriginalLanguageID: languageEnglish,
			CountryID:          countryUSA,
			PictureFileKey:     "img/3.jpg",
			Genres:             []string{"Боевик", "Криминал", "Драма"},
			Actors: []domain.ActorPreview{
				actorPreview(baleID, "Кристиан Бейл", "img/actors/bale.jpg"),
			},
		},
		{
			ID:                 prestigeID,
			Title:              "Престиж",
			Description:        "История соперничества двух выдающихся иллюзионистов.",
			Director:           "Кристофер Нолан",
			ContentType:        contentTypeFilm,
			ReleaseYear:        2006,
			DurationSeconds:    130 * 60,
			AgeLimit:           12,
			OriginalLanguageID: languageEnglish,
			CountryID:          countryUSA,
			PictureFileKey:     "img/4.jpg",
			Genres:             []string{"Драма", "Триллер", "Детектив"},
			Actors: []domain.ActorPreview{
				actorPreview(baleID, "Кристиан Бейл", "img/actors/bale.jpg"),
			},
		},
		{
			ID:                 fordFerrariID,
			Title:              "Ford против Ferrari",
			Description:        "История инженеров и гонщиков, создавших автомобиль, бросивший вызов Ferrari.",
			Director:           "Джеймс Мэнголд",
			ContentType:        contentTypeFilm,
			ReleaseYear:        2019,
			DurationSeconds:    152 * 60,
			AgeLimit:           12,
			OriginalLanguageID: languageEnglish,
			CountryID:          countryUSA,
			PictureFileKey:     "img/5.jpg",
			Genres:             []string{"Биография", "Драма", "Спорт"},
			Actors: []domain.ActorPreview{
				actorPreview(baleID, "Кристиан Бейл", "img/actors/bale.jpg"),
			},
		},
	}

	actors := []domain.ActorResponse{
		{
			ID:             chalametID,
			FullName:       "Тимати Шаламе",
			Biography:      "Американский актёр, известный ролями в драматических и фантастических фильмах.",
			BirthDate:      "1995-12-27",
			CountryID:      countryUSA,
			PictureFileKey: "img/actors/chalamet.jpg",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
			},
		},
		{
			ID:             zendayaID,
			FullName:       "Зендея",
			Biography:      "Американская актриса и певица.",
			BirthDate:      "1996-09-01",
			CountryID:      countryUSA,
			PictureFileKey: "img/actors/zendaya.jpg",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
			},
		},
		{
			ID:             bardemID,
			FullName:       "Хавьер Бардем",
			Biography:      "Испанский актёр, лауреат премии «Оскар».",
			BirthDate:      "1969-03-01",
			CountryID:      countrySpain,
			PictureFileKey: "img/actors/bardem.jpg",
			Movies: []domain.MoviePreview{
				moviePreview(duneID, "Дюна: Часть Вторая", "img/1.jpg"),
			},
		},
		{
			ID:             phoenixID,
			FullName:       "Хоакин Феникс",
			Biography:      "Американский актёр, известный интенсивными драматическими ролями.",
			BirthDate:      "1974-10-28",
			CountryID:      countryUSA,
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
			CountryID:      countryUK,
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

		if err := r.db.Save("movies", strconv.FormatInt(movie.ID, 10), movieData); err != nil {
			log.Println(err)
		}
	}

	for _, actor := range actors {
		actorData, err := serializer.Serialize(actor)
		if err != nil {
			log.Println(err)

			continue
		}

		if err := r.db.Save("actors", strconv.FormatInt(actor.ID, 10), actorData); err != nil {
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
