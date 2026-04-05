INSERT INTO country (title) VALUES
    ('США'),
    ('Испания'),
    ('Великобритания')
ON CONFLICT (title) DO NOTHING;

INSERT INTO language (title) VALUES
    ('Английский')
ON CONFLICT (title) DO NOTHING;

INSERT INTO genre (title) VALUES
    ('Фантастика'),
    ('Драма'),
    ('Приключения'),
    ('Триллер'),
    ('Боевик'),
    ('Криминал'),
    ('Детектив'),
    ('Биография'),
    ('Спорт')
ON CONFLICT (title) DO NOTHING;

INSERT INTO movie (
    title, description, director, content_type, release_year, 
    duration_seconds, age_limit, original_language_id, country_id, picture_file_key
) VALUES 
    (
        'Дюна: Часть Вторая', 
        'Пол Атрейдес, объединившись с фрименами, продолжает путь мести и принимает судьбоносные решения на Арракисе.',
        'Дени Вильнёв', 'film', 2024, 9960, 16, 
        (SELECT id FROM language WHERE title = 'Английский'), 
        (SELECT id FROM country WHERE title = 'США'), 
        'img/1.jpg'
    ),
    (
        'Джокер',
        'История Артура Флека, который постепенно превращается в Джокера.',
        'Тодд Филлипс', 'film', 2019, 7320, 18,
        (SELECT id FROM language WHERE title = 'Английский'), 
        (SELECT id FROM country WHERE title = 'США'), 
        'img/2.jpeg'
    ),
    (
        'Тёмный рыцарь',
        'Бэтмен сталкивается с Джокером, который погружает Готэм в хаос.',
        'Кристофер Нолан', 'film', 2008, 9120, 16,
        (SELECT id FROM language WHERE title = 'Английский'), 
        (SELECT id FROM country WHERE title = 'США'), 
        'img/3.jpg'
    ),
    (
        'Престиж',
        'История соперничества двух выдающихся иллюзионистов.',
        'Кристофер Нолан', 'film', 2006, 7800, 12,
        (SELECT id FROM language WHERE title = 'Английский'), 
        (SELECT id FROM country WHERE title = 'США'), 
        'img/4.jpg'
    ),
    (
        'Ford против Ferrari',
        'История инженеров и гонщиков, создавших автомобиль, бросивший вызов Ferrari.',
        'Джеймс Мэнголд', 'film', 2019, 9120, 12,
        (SELECT id FROM language WHERE title = 'Английский'), 
        (SELECT id FROM country WHERE title = 'США'), 
        'img/5.jpg'
    );

INSERT INTO genre_to_movie (genre_id, movie_id)
SELECT g.id, m.id
FROM movie m
CROSS JOIN genre g
WHERE (m.title = 'Дюна: Часть Вторая' AND g.title IN ('Фантастика', 'Драма', 'Приключения'))
   OR (m.title = 'Джокер' AND g.title IN ('Драма', 'Триллер'))
   OR (m.title = 'Тёмный рыцарь' AND g.title IN ('Боевик', 'Криминал', 'Драма'))
   OR (m.title = 'Престиж' AND g.title IN ('Драма', 'Триллер', 'Детектив'))
   OR (m.title = 'Ford против Ferrari' AND g.title IN ('Биография', 'Драма', 'Спорт'))
ON CONFLICT DO NOTHING;

INSERT INTO actor (full_name, birthdate, biography, country_id, picture_file_key) VALUES
    (
        'Тимати Шаламе', '1995-12-27',
        'Американский актёр, известный ролями в драматических и фантастических фильмах.',
        (SELECT id FROM country WHERE title = 'США'), 'img/actors/chalamet.jpg'
    ),
    (
        'Зендея', '1996-09-01',
        'Американская актриса и певица.',
        (SELECT id FROM country WHERE title = 'США'), 'img/actors/zendaya.jpg'
    ),
    (
        'Хавьер Бардем', '1969-03-01',
        'Испанский актёр, лауреат премии «Оскар».',
        (SELECT id FROM country WHERE title = 'Испания'), 'img/actors/bardem.jpg'
    ),
    (
        'Хоакин Феникс', '1974-10-28',
        'Американский актёр, известный интенсивными драматическими ролями.',
        (SELECT id FROM country WHERE title = 'США'), 'img/actors/phoenix.jpg'
    ),
    (
        'Кристиан Бейл', '1974-01-30',
        'Британский актёр, известный ролями в психологически и физически сложных образах.',
        (SELECT id FROM country WHERE title = 'Великобритания'), 'img/actors/bale.jpg'
    );

INSERT INTO actor_to_movie (actor_id, movie_id)
SELECT a.id, m.id
FROM movie m
CROSS JOIN actor a
WHERE (m.title = 'Дюна: Часть Вторая' AND a.full_name IN ('Тимати Шаламе', 'Зендея', 'Хавьер Бардем'))
   OR (m.title = 'Джокер' AND a.full_name = 'Хоакин Феникс')
   OR (m.title = 'Тёмный рыцарь' AND a.full_name = 'Кристиан Бейл')
   OR (m.title = 'Престиж' AND a.full_name = 'Кристиан Бейл')
   OR (m.title = 'Ford против Ferrari' AND a.full_name = 'Кристиан Бейл')
ON CONFLICT DO NOTHING;

INSERT INTO selection (title) VALUES
    ('Популярные'),
    ('Новинки'),
    ('Топ-10')
ON CONFLICT (title) DO NOTHING;

INSERT INTO movie_to_selection (movie_id, selection_id)
SELECT m.id, s.id
FROM movie m
CROSS JOIN selection s
WHERE (s.title = 'Популярные' AND m.title IN ('Дюна: Часть Вторая', 'Джокер', 'Тёмный рыцарь', 'Престиж', 'Ford против Ferrari'))
   OR (s.title = 'Новинки' AND m.title IN ('Дюна: Часть Вторая', 'Ford против Ferrari', 'Джокер'))
   OR (s.title = 'Топ-10' AND m.title IN ('Тёмный рыцарь', 'Престиж', 'Джокер'))
ON CONFLICT DO NOTHING;