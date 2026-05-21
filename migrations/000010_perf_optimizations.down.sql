BEGIN;

drop index if exists user_interaction_favorite_user_updated_idx;
drop index if exists user_interaction_movie_review_idx;
drop index if exists user_interaction_movie_rating_idx;
drop index if exists movie_to_selection_selection_id_id_movie_id_idx;
drop index if exists genre_to_movie_movie_id_genre_id_idx;
drop index if exists actor_to_movie_movie_id_actor_id_idx;

COMMIT;
