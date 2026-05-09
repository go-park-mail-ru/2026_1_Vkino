\set ON_ERROR_STOP on

delete from selection
where title like 'PerfTest Selection %';

delete from users
where email like 'perf_test_user_%@example.test';

delete from movie
where title like 'PerfTest Movie %';

delete from actor
where full_name like 'PerfTest Actor %';

analyze selection;
analyze users;
analyze movie;
analyze actor;
analyze user_interaction;
analyze user_interaction_review_reaction;
analyze movie_to_selection;
analyze actor_to_movie;
analyze genre_to_movie;
analyze watch_progress_episode;
analyze episode;
analyze movie_external_rating;
