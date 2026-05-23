BEGIN;

ALTER TABLE movie
    ADD COLUMN IF NOT EXISTS is_paid boolean NOT NULL DEFAULT false;

ALTER TABLE episode
    ADD COLUMN IF NOT EXISTS is_paid boolean NOT NULL DEFAULT false;

UPDATE movie
SET is_paid = false;

WITH ranked_episodes AS (
    SELECT
        e.id,
        m.content_type,
        row_number() OVER (
            PARTITION BY e.movie_id
            ORDER BY e.season_number, e.episode_number, e.id
        ) AS episode_rank
    FROM episode e
    JOIN movie m ON m.id = e.movie_id
)
UPDATE episode e
SET is_paid = CASE
    WHEN ranked_episodes.content_type <> 'series' THEN false
    WHEN ranked_episodes.episode_rank = 1 THEN false
    ELSE true
END
FROM ranked_episodes
WHERE ranked_episodes.id = e.id;

COMMIT;
