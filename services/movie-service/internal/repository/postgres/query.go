package postgres

const (
	sqlGetMovieBaseByID = `
		select
			m.id,
			m.title,
			coalesce(m.description, '') as description,
			m.release_year,
			m.age_limit,
			ceil(m.duration_seconds / 60.0)::int as duration_min,
			m.poster_file_key,
			m.picture_file_key as card_file_key
		from movie m
		where m.id = $1
	`

	sqlGetMovieCountriesByID = `
		select c.title
		from movie m
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
			a.full_name as name,
			a.picture_file_key as avatar_file_key
		from actor_to_movie am
		join actor a on a.id = am.actor_id
		where am.movie_id = $1
		order by a.full_name
	`

	sqlGetMovieEpisodesByID = `
		select
			e.id,
			e.movie_id,
			e.episode_number as number,
			coalesce(e.title, '') as title,
			e.duration_seconds as duration_sec,
			e.video_file_key
		from episode e
		where e.movie_id = $1
		order by e.season_number, e.episode_number, e.id
	`

	sqlGetActorBaseByID = `
		select
			a.id,
			a.full_name as name,
			coalesce(a.biography, '') as description,
			a.picture_file_key as avatar_file_key
		from actor a
		where a.id = $1
	`

	sqlGetActorMoviesByID = `
		select
			m.id,
			m.title,
			m.release_year,
			m.poster_file_key,
			m.picture_file_key as card_file_key
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
			m.release_year,
			m.poster_file_key,
			m.picture_file_key as card_file_key
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
			m.release_year,
			m.poster_file_key,
			m.picture_file_key as card_file_key
		from selection s
		join movie_to_selection ms on ms.selection_id = s.id
		join movie m on m.id = ms.movie_id
		order by s.id, ms.id
	`

	sqlSearchMovies = `
		select
			m.id,
			m.title,
			m.release_year,
			m.poster_file_key,
			m.picture_file_key as card_file_key
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
			e.episode_number as number,
			coalesce(e.title, '') as title,
			e.duration_seconds as duration_sec,
			e.video_file_key
		from episode e
		where e.id = $1
	`

	sqlGetEpisodeProgress = `
		select
			wpe.episode_id,
			wpe.position_seconds as position_sec
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
		returning episode_id, position_seconds as position_sec
	`
)