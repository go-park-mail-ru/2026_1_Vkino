\set ON_ERROR_STOP on

with load_user as (
    select id
    from users
    where email = :'load_user_email'
)
delete from watch_progress_episode
where user_id = (select id from load_user);

with load_user as (
    select id
    from users
    where email = :'load_user_email'
),
seed_watch_user as (
    select wpe.user_id
    from watch_progress_episode wpe
    group by wpe.user_id
    order by count(*) desc, wpe.user_id
    limit 1
)
insert into watch_progress_episode (
    user_id,
    episode_id,
    position_seconds,
    created_at,
    updated_at
)
select
    lu.id,
    wpe.episode_id,
    wpe.position_seconds,
    wpe.created_at,
    wpe.updated_at
from seed_watch_user swu
join watch_progress_episode wpe on wpe.user_id = swu.user_id
cross join load_user lu
on conflict (user_id, episode_id)
do update set
    position_seconds = excluded.position_seconds,
    updated_at = excluded.updated_at;

with load_user as (
    select id
    from users
    where email = :'load_user_email'
)
update user_interaction
set is_favorite = false
where user_id = (select id from load_user)
    and is_favorite = true;

with load_user as (
    select id
    from users
    where email = :'load_user_email'
),
seed_favorite_user as (
    select ui.user_id
    from user_interaction ui
    where ui.is_favorite = true
    group by ui.user_id
    order by count(*) desc, ui.user_id
    limit 1
)
insert into user_interaction (
    movie_id,
    user_id,
    rating,
    comment,
    is_favorite,
    created_at,
    updated_at
)
select
    ui.movie_id,
    lu.id,
    ui.rating,
    ui.comment,
    true,
    ui.created_at,
    ui.updated_at
from seed_favorite_user sfu
join user_interaction ui on ui.user_id = sfu.user_id
cross join load_user lu
where ui.is_favorite = true
on conflict (movie_id, user_id)
do update set
    is_favorite = true,
    rating = excluded.rating,
    comment = excluded.comment,
    updated_at = excluded.updated_at;
