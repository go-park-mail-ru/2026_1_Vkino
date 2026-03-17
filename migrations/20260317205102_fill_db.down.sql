BEGIN;

DELETE FROM friend
WHERE user1_id IN (SELECT id FROM users WHERE email IN ('demo1@vkino.local', 'demo2@vkino.local'))
   OR user2_id IN (SELECT id FROM users WHERE email IN ('demo1@vkino.local', 'demo2@vkino.local'));

DELETE FROM watch_progress_episode
WHERE user_id IN (SELECT id FROM users WHERE email IN ('demo1@vkino.local', 'demo2@vkino.local'));

DELETE FROM user_interaction
WHERE user_id IN (SELECT id FROM users WHERE email IN ('demo1@vkino.local', 'demo2@vkino.local'));

DELETE FROM users
WHERE email IN ('demo1@vkino.local', 'demo2@vkino.local');

DELETE FROM movie_to_selection
WHERE selection_id IN (
    SELECT id FROM selection WHERE title IN (
        'Вечер с напряжением',
        'Легендарные истории',
        'Что посмотреть за выходные'
    )
);

DELETE FROM selection
WHERE title IN ('Вечер с напряжением', 'Легендарные истории', 'Что посмотреть за выходные');

DELETE FROM actor_to_movie
WHERE actor_id IN (
    SELECT id FROM actor WHERE full_name IN (
        'Keanu Reeves',
        'Bryan Cranston',
        'Александр Петров',
        'Rumi Hiiragi'
    )
);

DELETE FROM genre_to_movie
WHERE movie_id IN (
    SELECT id FROM movie WHERE picture_file_key IN (
        'movies/the-matrix/poster.jpg',
        'movies/breaking-bad/poster.jpg',
        'movies/tekst/poster.jpg',
        'movies/spirited-away/poster.jpg'
    )
);

DELETE FROM episode
WHERE picture_file_key IN (
    'episodes/breaking-bad/s01e01/poster.jpg',
    'episodes/breaking-bad/s01e02/poster.jpg'
);

DELETE FROM movie
WHERE picture_file_key IN (
    'movies/the-matrix/poster.jpg',
    'movies/breaking-bad/poster.jpg',
    'movies/tekst/poster.jpg',
    'movies/spirited-away/poster.jpg'
);

DELETE FROM actor
WHERE full_name IN ('Keanu Reeves', 'Bryan Cranston', 'Александр Петров', 'Rumi Hiiragi');

DELETE FROM genre
WHERE title IN ('Drama', 'Sci-Fi', 'Comedy', 'Thriller', 'Animation');

DELETE FROM country
WHERE title IN ('USA', 'Russia', 'Japan');

DELETE FROM language
WHERE title IN ('English', 'Russian', 'Japanese');

COMMIT;
