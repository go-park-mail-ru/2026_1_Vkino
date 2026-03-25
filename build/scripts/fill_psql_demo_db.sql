BEGIN;

-- Languages
insert into language (title, description) values
    ('English', 'Английский язык'),
    ('Russian', 'Русский язык'),
    ('Japanese', 'Японский язык')
on conflict (title) do nothing;

-- Countries
insert into country (title) values
    ('USA'),
    ('Russia'),
    ('Japan')
on conflict (title) do nothing;

-- Genres
insert into genre (title) values
    ('Drama'),
    ('Sci-Fi'),
    ('Comedy'),
    ('Thriller'),
    ('Animation')
on conflict (title) do nothing;

-- Actors
insert into actor (full_name, birthday, biography) values
    ('Keanu Reeves', '1964-09-02', 'Canadian actor known for action and sci-fi roles.'),
    ('Bryan Cranston', '1956-03-07', 'American actor known for dramatic television roles.'),
    ('Александр Петров', '1989-01-25', 'Российский актёр театра и кино.'),
    ('Rumi Hiiragi', '1987-08-01', 'Japanese actress and voice actress.')
on conflict (full_name, birthday) do nothing;

-- Movies
insert into movie (
    title,
    description,
    content_type,
    release_year,
    duration_seconds,
    age_limit,
    original_language_id,
    country_id,
    picture_file_key
)
select * from (
    values
    (
        'The Matrix',
        'A hacker discovers the world is a simulation and joins the resistance.',
        'film',
        1999,
        8160,
        16,
        (select id from language where title = 'English'),
        (select id from country where title = 'USA'),
        'movies/the-matrix/poster.jpg'
    ),
    (
        'Breaking Bad',
        'A chemistry teacher turns to producing methamphetamine after a cancer diagnosis.',
        'series',
        2008,
        3000,
        18,
        (select id from language where title = 'English'),
        (select id from country where title = 'USA'),
        'movies/breaking-bad/poster.jpg'
    ),
    (
        'Текст',
        'Драматический триллер о последствиях одного сообщения в чужом телефоне.',
        'film',
        2019,
        7920,
        18,
        (select id from language where title = 'Russian'),
        (select id from country where title = 'Russia'),
        'movies/tekst/poster.jpg'
    ),
    (
        'Spirited Away',
        'A girl enters a world of spirits and must save her parents.',
        'film',
        2001,
        7500,
        12,
        (select id from language where title = 'Japanese'),
        (select id from country where title = 'Japan'),
        'movies/spirited-away/poster.jpg'
    )
) as v(
    title,
    description,
    content_type,
    release_year,
    duration_seconds,
    age_limit,
    original_language_id,
    country_id,
    picture_file_key
)
on conflict (picture_file_key) do nothing;

-- Episodes for the series Breaking Bad
insert into episode (
    movie_id,
    description,
    season_number,
    episode_number,
    title,
    duration_seconds,
    picture_file_key,
    video_file_key
)
select * from (
    values
    (
        (select id from movie where picture_file_key = 'movies/breaking-bad/poster.jpg'),
        'Walter White receives life-changing news.',
        1,
        1,
        'Pilot',
        3480,
        'episodes/breaking-bad/s01e01/poster.jpg',
        'episodes/breaking-bad/s01e01/video.mp4'
    ),
    (
        (select id from movie where picture_file_key = 'movies/breaking-bad/poster.jpg'),
        'Walt and Jesse deal with the aftermath of their first cook.',
        1,
        2,
        'Cat''s in the Bag...',
        1 * 3540,
        'episodes/breaking-bad/s01e02/poster.jpg',
        'episodes/breaking-bad/s01e02/video.mp4'
    )
) as v(
    movie_id,
    description,
    season_number,
    episode_number,
    title,
    duration_seconds,
    picture_file_key,
    video_file_key
)
on conflict (movie_id, season_number, episode_number) do nothing;

-- Genre links
insert into genre_to_movie (genre_id, movie_id)
select g.id, m.id
from (values
    ('Sci-Fi', 'The Matrix'),
    ('Thriller', 'The Matrix'),
    ('Drama', 'Breaking Bad'),
    ('Thriller', 'Breaking Bad'),
    ('Drama', 'Текст'),
    ('Thriller', 'Текст'),
    ('Animation', 'Spirited Away'),
    ('Drama', 'Spirited Away')
) as x(genre_title, movie_title)
join genre g on g.title = x.genre_title
join movie m on m.title = x.movie_title
on conflict (genre_id, movie_id) do nothing;

-- Actor links
insert into actor_to_movie (actor_id, movie_id)
select a.id, m.id
from (values
    ('Keanu Reeves', 'The Matrix'),
    ('Bryan Cranston', 'Breaking Bad'),
    ('Александр Петров', 'Текст'),
    ('Rumi Hiiragi', 'Spirited Away')
) as x(actor_name, movie_title)
join actor a on a.full_name = x.actor_name
join movie m on m.title = x.movie_title
on conflict (actor_id, movie_id) do nothing;

-- Selections
insert into selection (title, emotion, rating) values
    ('Вечер с напряжением', 'tense', 8.70),
    ('Легендарные истории', 'inspired', 9.10),
    ('Что посмотреть за выходные', 'calm', 8.40)
on conflict (title) do nothing;

-- Movie to selections
insert into movie_to_selection (movie_id, selection_id)
select m.id, s.id
from (values
    ('The Matrix', 'Легендарные истории'),
    ('Breaking Bad', 'Легендарные истории'),
    ('Текст', 'Вечер с напряжением'),
    ('The Matrix', 'Вечер с напряжением'),
    ('Spirited Away', 'Что посмотреть за выходные'),
    ('The Matrix', 'Что посмотреть за выходные')
) as x(movie_title, selection_title)
join movie m on m.title = x.movie_title
join selection s on s.title = x.selection_title
on conflict (movie_id, selection_id) do nothing;

-- Users
insert into users (email, password_hash, is_active) values
    ('demo1@vkino.local', '$2a$10$7EqJtq98hPqEX7fNZaFWoOHi6K4QvOB0fOkH1ZZ1xd6QbaO5jM90K', true),
    ('demo2@vkino.local', '$2a$10$7EqJtq98hPqEX7fNZaFWoOHi6K4QvOB0fOkH1ZZ1xd6QbaO5jM90K', true)
on conflict (email) do nothing;

-- Interactions
insert into user_interaction (movie_id, user_id, rating, is_favorite)
select m.id, u.id, x.rating, x.is_favorite
from (values
    ('The Matrix', 'demo1@vkino.local', 9.30::numeric(4,2), true),
    ('Breaking Bad', 'demo1@vkino.local', 9.80::numeric(4,2), true),
    ('Spirited Away', 'demo2@vkino.local', 8.90::numeric(4,2), false)
) as x(movie_title, email, rating, is_favorite)
join movie m on m.title = x.movie_title
join users u on u.email = x.email
on conflict (movie_id, user_id) do nothing;

-- Watch progress
insert into watch_progress_episode (user_id, episode_id, position_seconds)
select u.id, e.id, x.position_seconds
from (values
    ('demo1@vkino.local', 1, 1200),
    ('demo2@vkino.local', 2, 300)
) as x(email, episode_number, position_seconds)
join users u on u.email = x.email
join movie m on m.title = 'Breaking Bad'
join episode e on e.movie_id = m.id and e.season_number = 1 and e.episode_number = x.episode_number
on conflict (user_id, episode_id) do nothing;

-- Friends
insert into friend (user1_id, user2_id)
select least(u1.id, u2.id), greatest(u1.id, u2.id)
from users u1
join users u2 on u1.email = 'demo1@vkino.local' and u2.email = 'demo2@vkino.local'
on conflict (user1_id, user2_id) do nothing;

COMMIT;
