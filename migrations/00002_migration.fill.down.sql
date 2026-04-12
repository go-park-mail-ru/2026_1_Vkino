BEGIN;

DELETE FROM watch_progress_episode;
DELETE FROM episode;
DELETE FROM movie_to_selection;
DELETE FROM selection;
DELETE FROM actor_to_movie;
DELETE FROM actor;
DELETE FROM genre_to_movie;
DELETE FROM movie;
DELETE FROM genre;
DELETE FROM language;
DELETE FROM country;

COMMIT;
