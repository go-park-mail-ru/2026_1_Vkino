\set ON_ERROR_STOP on

set statement_timeout = 0;

begin;

create temp table perf_genre_ids on commit drop as
select row_number() over (order by g.id) as seq, g.id
from genre g;

create temp table perf_country_ids on commit drop as
select row_number() over (order by c.id) as seq, c.id
from country c;

create temp table perf_language_ids on commit drop as
select row_number() over (order by l.id) as seq, l.id
from language l;

create temp table perf_user_ids (
    seq int primary key,
    user_id bigint not null
) on commit drop;

with src as (
    select
        gs as seq,
        format('perf_test_user_%s@example.test', lpad(gs::text, 6, '0')) as email,
        '$2y$10$gt5iYQOBFpuPgjGVW1v1duj8jCEx.8Q8OOcxT.LUNOSBQAi7w0XT.' as password_hash,
        date '1980-01-01' + ((gs * 17) % 12000) as birthdate,
        format('perf_test/user/%s/avatar.webp', lpad(gs::text, 6, '0')) as avatar_file_key,
        timestamptz '2026-01-01 00:00:00+00' + make_interval(days => ((gs % 120)::int)) as registration_date,
        true as is_active
    from generate_series(1, :user_count) as gs
),
ins as (
    insert into users (
        email,
        password_hash,
        birthdate,
        avatar_file_key,
        registration_date,
        is_active,
        role,
        created_at,
        updated_at
    )
    select
        s.email,
        s.password_hash,
        s.birthdate,
        s.avatar_file_key,
        s.registration_date,
        s.is_active,
        'user',
        s.registration_date,
        s.registration_date
    from src s
    returning id, email
)
insert into perf_user_ids (seq, user_id)
select row_number() over (order by i.email), i.id
from ins i
order by i.email;

create temp table perf_actor_ids (
    seq int primary key,
    actor_id bigint not null
) on commit drop;

with src as (
    select
        gs as seq,
        format('PerfTest Actor %s', lpad(gs::text, 6, '0')) as full_name,
        date '1950-01-01' + ((gs * 23) % 22000) as birthdate,
        format(
            'PerfTest biography for actor %s focused on drama, thrillers and long-running franchise work.',
            lpad(gs::text, 6, '0')
        ) as biography,
        ((gs - 1) % (select count(*) from perf_country_ids)) + 1 as country_seq,
        format('perf_test/actor/%s/picture.webp', lpad(gs::text, 6, '0')) as picture_file_key,
        timestamptz '2026-01-01 00:00:00+00' + make_interval(days => ((gs % 90)::int)) as created_at
    from generate_series(1, :actor_count) as gs
),
ins as (
    insert into actor (
        full_name,
        birthdate,
        biography,
        country_id,
        picture_file_key,
        created_at,
        updated_at
    )
    select
        s.full_name,
        s.birthdate,
        s.biography,
        c.id,
        s.picture_file_key,
        s.created_at,
        s.created_at
    from src s
    join perf_country_ids c on c.seq = s.country_seq
    returning id, full_name
)
insert into perf_actor_ids (seq, actor_id)
select row_number() over (order by i.full_name), i.id
from ins i
order by i.full_name;

create temp table perf_movie_ids (
    seq int primary key,
    movie_id bigint not null,
    content_type text not null
) on commit drop;

with src as (
    select
        gs as seq,
        format('PerfTest Movie %s', lpad(gs::text, 6, '0')) as title,
        format(
            'PerfTest synopsis %s: detective sci-fi drama with recurring hero arcs, neon megacities and family conflicts.',
            lpad(gs::text, 6, '0')
        ) as description,
        format('Perf Director %s', lpad((((gs - 1) % 900) + 1)::text, 4, '0')) as director,
        case when gs % 5 = 0 then 'series' else 'film' end as content_type,
        1980 + (gs % 45) as release_year,
        case when gs % 5 = 0 then 2700 else 4800 + ((gs * 37) % 4200) end as duration_seconds,
        (array[0, 6, 12, 14, 16, 18])[((gs - 1) % 6) + 1] as age_limit,
        ((gs - 1) % (select count(*) from perf_language_ids)) + 1 as language_seq,
        ((gs - 1) % (select count(*) from perf_country_ids)) + 1 as country_seq,
        format('perf_test/movie/%s/card.webp', lpad(gs::text, 6, '0')) as picture_file_key,
        format('perf_test/movie/%s/poster.webp', lpad(gs::text, 6, '0')) as poster_file_key,
        timestamptz '2026-01-01 00:00:00+00' + make_interval(days => ((gs % 180)::int)) as created_at
    from generate_series(1, :movie_count) as gs
),
ins as (
    insert into movie (
        title,
        description,
        director,
        content_type,
        release_year,
        duration_seconds,
        age_limit,
        original_language_id,
        country_id,
        picture_file_key,
        poster_file_key,
        created_at,
        updated_at
    )
    select
        s.title,
        s.description,
        s.director,
        s.content_type,
        s.release_year,
        s.duration_seconds,
        s.age_limit,
        l.id,
        c.id,
        s.picture_file_key,
        s.poster_file_key,
        s.created_at,
        s.created_at
    from src s
    join perf_language_ids l on l.seq = s.language_seq
    join perf_country_ids c on c.seq = s.country_seq
    returning id, title, content_type
)
insert into perf_movie_ids (seq, movie_id, content_type)
select row_number() over (order by i.title), i.id, i.content_type
from ins i
order by i.title;

insert into genre_to_movie (genre_id, movie_id, created_at, updated_at)
select
    g.id,
    pm.movie_id,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => ((pm.seq % 90)::int)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => ((pm.seq % 90)::int))
from perf_movie_ids pm
join perf_genre_ids g on g.seq = ((pm.seq - 1) % (select count(*) from perf_genre_ids)) + 1
union all
select
    g.id,
    pm.movie_id,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + 1) % 90)::int)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + 1) % 90)::int))
from perf_movie_ids pm
join perf_genre_ids g on g.seq = (((pm.seq + 2) - 1) % (select count(*) from perf_genre_ids)) + 1
on conflict do nothing;

insert into actor_to_movie (actor_id, movie_id, created_at, updated_at)
select
    pa.actor_id,
    pm.movie_id,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => ((pm.seq % 75)::int)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => ((pm.seq % 75)::int))
from perf_movie_ids pm
join perf_actor_ids pa on pa.seq = ((pm.seq * 7) % :actor_count) + 1
union all
select
    pa.actor_id,
    pm.movie_id,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + 1) % 75)::int)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + 1) % 75)::int))
from perf_movie_ids pm
join perf_actor_ids pa on pa.seq = ((pm.seq * 11) % :actor_count) + 1
union all
select
    pa.actor_id,
    pm.movie_id,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + 2) % 75)::int)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + 2) % 75)::int))
from perf_movie_ids pm
join perf_actor_ids pa on pa.seq = ((pm.seq * 13) % :actor_count) + 1
on conflict do nothing;

create temp table perf_series_movies on commit drop as
select row_number() over (order by movie_id) as seq, movie_id
from perf_movie_ids
where content_type = 'series';

create temp table perf_episode_ids (
    seq bigint primary key,
    episode_id bigint not null,
    duration_seconds int not null
) on commit drop;

with src as (
    select
        sm.movie_id,
        episode_no,
        ((episode_no - 1) / 3) + 1 as season_number,
        ((episode_no - 1) % 3) + 1 as episode_number,
        format('PerfTest Episode %s', episode_no) as title,
        format('PerfTest episode description for movie %s episode %s.', sm.movie_id, episode_no) as description,
        2100 + ((sm.seq * 17 + episode_no * 29) % 1800) as duration_seconds,
        format('perf_test/episode/%s/%s/card.webp', sm.movie_id, episode_no) as picture_file_key,
        format('perf_test/episode/%s/%s/video.mp4', sm.movie_id, episode_no) as video_file_key,
        timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((sm.seq + episode_no) % 150)::int)) as created_at
    from perf_series_movies sm
    cross join generate_series(1, 6) as episode_no
),
ins as (
    insert into episode (
        movie_id,
        description,
        season_number,
        episode_number,
        title,
        duration_seconds,
        picture_file_key,
        video_file_key,
        created_at,
        updated_at
    )
    select
        s.movie_id,
        s.description,
        s.season_number,
        s.episode_number,
        s.title,
        s.duration_seconds,
        s.picture_file_key,
        s.video_file_key,
        s.created_at,
        s.created_at
    from src s
    returning id, movie_id, season_number, episode_number, duration_seconds
)
insert into perf_episode_ids (seq, episode_id, duration_seconds)
select
    row_number() over (order by i.movie_id, i.season_number, i.episode_number),
    i.id,
    i.duration_seconds
from ins i
order by i.movie_id, i.season_number, i.episode_number;

create temp table perf_selection_ids (
    seq int primary key,
    selection_id bigint not null
) on commit drop;

with src as (
    select
        gs as seq,
        format('PerfTest Selection %s', lpad(gs::text, 3, '0')) as title,
        format('Emotion %s', lpad(gs::text, 3, '0')) as emotion,
        timestamptz '2026-01-01 00:00:00+00' + make_interval(days => ((gs % 60)::int)) as created_at
    from generate_series(1, :selection_count) as gs
),
ins as (
    insert into selection (title, emotion, rating, created_at, updated_at)
    select
        s.title,
        s.emotion,
        null,
        s.created_at,
        s.created_at
    from src s
    returning id, title
)
insert into perf_selection_ids (seq, selection_id)
select row_number() over (order by i.title), i.id
from ins i
order by i.title;

insert into movie_to_selection (movie_id, selection_id, created_at, updated_at)
select
    pm.movie_id,
    ps.selection_id,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((ps.seq + slot) % 45)::int)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((ps.seq + slot) % 45)::int))
from perf_selection_ids ps
cross join generate_series(1, :movies_per_selection) as slot
join perf_movie_ids pm on pm.seq = (((ps.seq - 1) * :movies_per_selection + slot * 37) % :movie_count) + 1
on conflict do nothing;

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
    pm.movie_id,
    pu.user_id,
    case
        when (pm.seq + slot) % 9 = 0 then null
        else round(((((pm.seq * 17 + slot * 13) % 91) + 10)::numeric / 10), 2)
    end as rating,
    case
        when slot % 4 = 0 then format(
            'PerfTest review for movie %s slot %s: pacing, acting and visual style are discussed in detail.',
            pm.seq,
            slot
        )
        else null
    end as comment,
    ((pm.seq + slot) % 7 = 0) as is_favorite,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + slot) % 210)::int), mins => (slot * 3)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq * slot) % 210)::int), mins => (slot * 11))
from perf_movie_ids pm
cross join generate_series(1, :interactions_per_movie) as slot
join perf_user_ids pu on pu.seq = ((pm.seq * 37 + slot * 101) % :user_count) + 1
on conflict (movie_id, user_id)
do nothing;

create temp table perf_review_ids on commit drop as
select
    row_number() over (order by ui.id) as seq,
    ui.id as review_id,
    ui.user_id as review_author_id
from user_interaction ui
where ui.comment like 'PerfTest review%';

insert into user_interaction_review_reaction (
    review_id,
    user_id,
    reaction,
    created_at,
    updated_at
)
select
    pr.review_id,
    pu.user_id,
    case when (pr.seq + slot) % 5 = 0 then 'dislike' else 'like' end,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pr.seq + slot) % 120)::int), mins => slot),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pr.seq + slot) % 120)::int), mins => slot)
from perf_review_ids pr
cross join generate_series(1, 2) as slot
join perf_user_ids pu on pu.seq = ((pr.seq * 53 + slot * 97) % :user_count) + 1
where pu.user_id <> pr.review_author_id
on conflict (review_id, user_id)
do nothing;

insert into movie_external_rating (
    movie_id,
    source,
    value,
    scale,
    created_at,
    updated_at
)
select
    pm.movie_id,
    src.source,
    src.value,
    10.0,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + src.src_order) % 90)::int)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pm.seq + src.src_order) % 90)::int))
from perf_movie_ids pm
join lateral (
    values
        (1, 'imdb', round(((((pm.seq * 19) % 35) + 60)::numeric / 10), 2)),
        (2, 'kinopoisk', round(((((pm.seq * 23) % 30) + 65)::numeric / 10), 2))
) as src(src_order, source, value) on true
where pm.seq % 2 = 0
on conflict (movie_id, source)
do nothing;

insert into watch_progress_episode (
    user_id,
    episode_id,
    position_seconds,
    created_at,
    updated_at
)
select
    pu.user_id,
    pe.episode_id,
    case
        when slot % 5 = 0 then greatest(0, pe.duration_seconds * 97 / 100)
        when slot % 3 = 0 then greatest(0, pe.duration_seconds * 55 / 100)
        else greatest(0, pe.duration_seconds * 20 / 100)
    end as position_seconds,
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pu.seq + slot) % 180)::int), mins => (slot * 5)),
    timestamptz '2026-01-01 00:00:00+00' + make_interval(days => (((pu.seq * slot) % 180)::int), mins => (slot * 13))
from perf_user_ids pu
cross join generate_series(1, :watch_rows_per_user) as slot
join perf_episode_ids pe on pe.seq = ((pu.seq * 11 + slot * 97) % (select count(*) from perf_episode_ids)) + 1
on conflict (user_id, episode_id)
do update set
    position_seconds = excluded.position_seconds,
    updated_at = excluded.updated_at;

commit;

analyze users;
analyze actor;
analyze movie;
analyze genre_to_movie;
analyze actor_to_movie;
analyze episode;
analyze selection;
analyze movie_to_selection;
analyze user_interaction;
analyze user_interaction_review_reaction;
analyze movie_external_rating;
analyze watch_progress_episode;

select 'perf_test_summary' as marker, 'movies' as entity, count(*)::bigint as rows_count
from movie
where title like 'PerfTest Movie %'
union all
select 'perf_test_summary', 'actors', count(*)::bigint
from actor
where full_name like 'PerfTest Actor %'
union all
select 'perf_test_summary', 'selections', count(*)::bigint
from selection
where title like 'PerfTest Selection %'
union all
select 'perf_test_summary', 'users', count(*)::bigint
from users
where email like 'perf_test_user_%@example.test'
union all
select 'perf_test_summary', 'user_interaction', count(*)::bigint
from user_interaction ui
join movie m on m.id = ui.movie_id
where m.title like 'PerfTest Movie %'
union all
select 'perf_test_summary', 'review_reaction', count(*)::bigint
from user_interaction_review_reaction uirr
join user_interaction ui on ui.id = uirr.review_id
join movie m on m.id = ui.movie_id
where m.title like 'PerfTest Movie %'
union all
select 'perf_test_summary', 'episodes', count(*)::bigint
from episode e
join movie m on m.id = e.movie_id
where m.title like 'PerfTest Movie %'
union all
select 'perf_test_summary', 'watch_progress', count(*)::bigint
from watch_progress_episode wpe
join users u on u.id = wpe.user_id
where u.email like 'perf_test_user_%@example.test';
