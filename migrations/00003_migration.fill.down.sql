BEGIN;

DELETE FROM episode;
DELETE FROM actor_to_movie;
DELETE FROM actor;

COMMIT;
