set statement_timeout = 0;

BEGIN;

create index if not exists actor_to_movie_movie_id_actor_id_idx
    on actor_to_movie (movie_id, actor_id);

create index if not exists genre_to_movie_movie_id_genre_id_idx
    on genre_to_movie (movie_id, genre_id);

create index if not exists movie_to_selection_selection_id_id_movie_id_idx
    on movie_to_selection (selection_id, id, movie_id);

create index if not exists user_interaction_movie_rating_idx
    on user_interaction (movie_id)
    include (rating)
    where rating is not null;

create index if not exists user_interaction_movie_review_idx
    on user_interaction (movie_id, updated_at desc, id desc)
    include (user_id, rating, comment, created_at)
    where rating is not null or nullif(btrim(coalesce(comment, '')), '') is not null;

create index if not exists user_interaction_favorite_user_updated_idx
    on user_interaction (user_id, updated_at desc)
    include (movie_id)
    where is_favorite = true;

COMMIT;
