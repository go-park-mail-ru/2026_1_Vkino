BEGIN;

INSERT INTO actor_to_movie (actor_id, movie_id)
SELECT a.id, m.id
FROM actor a
         CROSS JOIN movie m
    ON CONFLICT (actor_id, movie_id) DO NOTHING;

COMMIT;