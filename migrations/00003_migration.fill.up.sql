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
    )
ON CONFLICT (full_name, birthdate) DO NOTHING;

INSERT INTO actor_to_movie (actor_id, movie_id)
SELECT a.id, m.id
FROM movie m
CROSS JOIN actor a
ON CONFLICT (actor_id, movie_id) DO NOTHING;

INSERT INTO episode (
    movie_id, description, season_number, episode_number, title,
    duration_seconds, picture_file_key, video_file_key
)
SELECT
    m.id,
    m.description,
    0,
    1,
    m.title,
    m.duration_seconds,
    m.picture_file_key,
    CASE m.title
        WHEN 'Дюна: Часть Вторая' THEN 'video/1.mp4'
        WHEN 'Джокер' THEN 'video/2.mp4'
        WHEN 'Тёмный рыцарь' THEN 'video/3.mp4'
        WHEN 'Престиж' THEN 'video/4.mp4'
        WHEN 'Ford против Ferrari' THEN 'video/5.mp4'
    END
FROM movie m
WHERE m.title IN (
    'Дюна: Часть Вторая',
    'Джокер',
    'Тёмный рыцарь',
    'Престиж',
    'Ford против Ferrari'
)
ON CONFLICT (movie_id, season_number, episode_number) DO NOTHING;
