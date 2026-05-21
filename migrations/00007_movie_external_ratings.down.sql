BEGIN;

delete from users
where email like '%@seed.vkino.local';

drop table if exists user_interaction_review_reaction;
alter table if exists user_interaction
    drop column if exists comment;
drop table if exists movie_external_rating;

COMMIT;
