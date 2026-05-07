package postgres

const (
	sqlGetMovieBaseByID = `
		select
			m.id,
			m.title,
			coalesce(m.description, ''),
			coalesce(m.director, ''),
			coalesce(m.trailer_url, ''),
			m.content_type,
			m.release_year,
			m.duration_seconds,
			m.age_limit,
			m.original_language_id,
			l.title,
			m.country_id,
			c.title,
			m.picture_file_key,
			coalesce(m.poster_file_key, '')
		from movie m
		join language l on l.id = m.original_language_id
		join country c on c.id = m.country_id
		where m.id = $1
	`

	sqlGetMovieGenresByID = `
		select g.title
		from genre_to_movie gm
		join genre g on g.id = gm.genre_id
		where gm.movie_id = $1
		order by g.title
	`

	sqlGetMovieActorsByID = `
		select
			a.id,
			a.full_name,
			a.picture_file_key
		from actor_to_movie am
		join actor a on a.id = am.actor_id
		where am.movie_id = $1
		order by a.full_name
	`

	sqlGetMovieEpisodesByID = `
		select
			e.id,
			e.movie_id,
			e.season_number,
			e.episode_number,
			coalesce(e.title, ''),
			coalesce(e.description, ''),
			e.duration_seconds,
			e.picture_file_key,
			e.video_file_key
		from episode e
		where e.movie_id = $1
		order by e.season_number, e.episode_number, e.id
	`

	sqlGetMovieExternalRatingsByID = `
		select
			mer.source,
			mer.value::double precision,
			mer.scale::double precision
		from movie_external_rating mer
		where mer.movie_id = $1
		order by mer.source
	`

	sqlGetMovieReviewsByMovieID = `
		with reaction_counts as (
			select
				uir.review_id,
				count(*) filter (where uir.reaction = 'like') as likes_count,
				count(*) filter (where uir.reaction = 'dislike') as dislikes_count
			from user_interaction_review_reaction uir
			group by uir.review_id
		),
		viewer_reactions as (
			select
				uir.review_id,
				uir.reaction
			from user_interaction_review_reaction uir
			where uir.user_id = $2
		)
		select
			ui.id,
			ui.user_id,
			u.email,
			ui.rating::double precision,
			coalesce(ui.comment, ''),
			coalesce(rc.likes_count, 0),
			coalesce(rc.dislikes_count, 0),
			coalesce(vr.reaction, ''),
			ui.created_at,
			ui.updated_at
		from user_interaction ui
		join users u on u.id = ui.user_id
		left join reaction_counts rc on rc.review_id = ui.id
		left join viewer_reactions vr on vr.review_id = ui.id
		where ui.movie_id = $1
			and (
				ui.rating is not null
				or nullif(btrim(coalesce(ui.comment, '')), '') is not null
			)
		order by ui.updated_at desc, ui.id desc
	`

	sqlGetActorBaseByID = `
		select
			a.id,
			a.full_name,
			a.birthdate,
			coalesce(a.biography, ''),
			a.country_id,
			c.title,
			a.picture_file_key
		from actor a
		join country c on c.id = a.country_id
		where a.id = $1
	`

	sqlGetActorMoviesByID = `
		select
			m.id,
			m.title,
			m.picture_file_key
		from actor_to_movie am
		join movie m on m.id = am.movie_id
		where am.actor_id = $1
		order by m.release_year desc, m.title
	`

	sqlGetGenreBaseByID = `
		select
			g.id,
			g.title
		from genre g
		where g.id = $1
	`

	sqlGetAllGenres = `
		select
			g.id,
			g.title
		from genre g
		order by g.title, g.id
	`

	sqlGetGenreMoviesByID = `
		select
			m.id,
			m.title,
			m.picture_file_key
		from genre_to_movie gm
		join movie m on m.id = gm.movie_id
		where gm.genre_id = $1
		order by m.release_year desc, m.title
	`

	sqlGetSelectionMoviesByTitle = `
		with movie_user_ratings as (
			select
				ui.movie_id,
				avg(ui.rating)::double precision as avg_rating
			from user_interaction ui
			where ui.rating is not null
			group by ui.movie_id
		),
		selection_ratings as (
			select
				ms.selection_id,
				round(avg(mur.avg_rating)::numeric, 2)::double precision as rating
			from movie_to_selection ms
			left join movie_user_ratings mur on mur.movie_id = ms.movie_id
			group by ms.selection_id
		)
		select
			s.title,
			sr.rating,
			m.id,
			m.title,
			m.picture_file_key
		from selection s
		join movie_to_selection ms on ms.selection_id = s.id
		join movie m on m.id = ms.movie_id
		left join selection_ratings sr on sr.selection_id = s.id
		where s.title = $1
		order by ms.id
	`

	sqlGetAllSelectionMovies = `
		with movie_user_ratings as (
			select
				ui.movie_id,
				avg(ui.rating)::double precision as avg_rating
			from user_interaction ui
			where ui.rating is not null
			group by ui.movie_id
		),
		selection_ratings as (
			select
				ms.selection_id,
				round(avg(mur.avg_rating)::numeric, 2)::double precision as rating
			from movie_to_selection ms
			left join movie_user_ratings mur on mur.movie_id = ms.movie_id
			group by ms.selection_id
		)
		select
			s.title,
			sr.rating,
			m.id,
			m.title,
			m.picture_file_key
		from selection s
		join movie_to_selection ms on ms.selection_id = s.id
		join movie m on m.id = ms.movie_id
		left join selection_ratings sr on sr.selection_id = s.id
		order by s.id, ms.id
	`

	sqlSearchMovies = `
		select
			m.id,
			m.title,
			m.picture_file_key
		from movie m
		where (
			setweight(to_tsvector('simple', coalesce(m.title, '')), 'A') ||
			setweight(to_tsvector('simple', coalesce(m.description, '')), 'B') ||
			setweight(to_tsvector('simple', coalesce(m.director, '')), 'C')
		) @@ to_tsquery('simple', $1)
		order by
			ts_rank(
				(
					setweight(to_tsvector('simple', coalesce(m.title, '')), 'A') ||
					setweight(to_tsvector('simple', coalesce(m.description, '')), 'B') ||
					setweight(to_tsvector('simple', coalesce(m.director, '')), 'C')
				),
				to_tsquery('simple', $1)
			) desc,
			m.release_year desc,
			m.title
		limit 50
	`

	sqlSearchActors = `
		select
			a.id,
			a.full_name,
			a.picture_file_key
		from actor a
		where (
			setweight(to_tsvector('simple', coalesce(a.full_name, '')), 'A') ||
			setweight(to_tsvector('simple', coalesce(a.biography, '')), 'B')
		) @@ to_tsquery('simple', $1)
		order by
			ts_rank(
				(
					setweight(to_tsvector('simple', coalesce(a.full_name, '')), 'A') ||
					setweight(to_tsvector('simple', coalesce(a.biography, '')), 'B')
				),
				to_tsquery('simple', $1)
			) desc,
			a.full_name
		limit 50
	`

	sqlGetMovieCardsByIDs = `
		select
			m.id,
			m.title,
			m.picture_file_key
		from movie m
		where m.id = any($1)
		order by array_position($1, m.id)
	`

	sqlGetEpisodePlayback = `
		select
			e.id,
			e.movie_id,
			e.season_number,
			e.episode_number,
			coalesce(e.title, ''),
			e.duration_seconds,
			e.video_file_key
		from episode e
		where e.id = $1
	`

	sqlGetEpisodeProgress = `
		select
			wpe.episode_id,
			coalesce(wpe.position_seconds, 0)
		from watch_progress_episode wpe
		where wpe.user_id = $1 and wpe.episode_id = $2
	`

	sqlSaveEpisodeProgress = `
		insert into watch_progress_episode (user_id, episode_id, position_seconds)
		values ($1, $2, $3)
		on conflict (user_id, episode_id)
		do update set
			position_seconds = excluded.position_seconds,
			updated_at = now()
		returning episode_id, position_seconds
	`

	sqlIsFavorite = `
		select exists(
			select 1 from user_interaction
			where user_id = $1 and movie_id = $2 and is_favorite = true
		)
	`

	sqlGetContinueWatching = `
		select
			wpe.episode_id, m.id, m.title, m.picture_file_key, m.content_type,
			e.season_number, e.episode_number, coalesce(e.title, m.title),
			wpe.position_seconds, e.duration_seconds, wpe.updated_at
		from watch_progress_episode wpe
		join episode e on e.id = wpe.episode_id
		join movie m on m.id = e.movie_id
		where wpe.user_id = $1
			and wpe.position_seconds < (e.duration_seconds * 0.9)
			and (CAST($3 AS double precision) <= 0)
		order by wpe.updated_at desc
		limit $2
	`

	sqlGetWatchHistory = `
		select
			wpe.episode_id, m.id, m.title, m.picture_file_key, m.content_type,
			e.season_number, e.episode_number, coalesce(e.title, m.title),
			wpe.position_seconds, e.duration_seconds, wpe.updated_at
		from watch_progress_episode wpe
		join episode e on e.id = wpe.episode_id
		join movie m on m.id = e.movie_id
		where wpe.user_id = $1
			and (
				CAST($3 AS double precision) <= 0
				or (
					e.duration_seconds > 0
					and wpe.position_seconds >= (
						e.duration_seconds * CAST($3 AS double precision)
					)
				)
			)
		order by wpe.updated_at desc
		limit $2
	`
)
