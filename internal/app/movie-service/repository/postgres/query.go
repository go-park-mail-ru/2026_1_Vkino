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

	sqlGetActorBaseByID = `
		select
			a.id,
			a.full_name,
			a.birthdate,
			coalesce(a.biography, ''),
			a.country_id,
			a.picture_file_key
		from actor a
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

	sqlGetSelectionMoviesByTitle = `
		select
			s.title,
			m.id,
			m.title,
			m.picture_file_key
		from selection s
		join movie_to_selection ms on ms.selection_id = s.id
		join movie m on m.id = ms.movie_id
		where s.title = $1
		order by ms.id
	`

	sqlGetAllSelectionMovies = `
		select
			s.title,
			m.id,
			m.title,
			m.picture_file_key
		from selection s
		join movie_to_selection ms on ms.selection_id = s.id
		join movie m on m.id = ms.movie_id
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
		) @@ plainto_tsquery('simple', $1)
		order by
			ts_rank(
				(
					setweight(to_tsvector('simple', coalesce(m.title, '')), 'A') ||
					setweight(to_tsvector('simple', coalesce(m.description, '')), 'B') ||
					setweight(to_tsvector('simple', coalesce(m.director, '')), 'C')
				),
				plainto_tsquery('simple', $1)
			) desc,
			m.release_year desc,
			m.title
		limit 50
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