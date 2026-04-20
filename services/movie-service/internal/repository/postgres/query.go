package postgres

const (
	sqlGetMovieBaseByID = `
		select
			m.id,
			m.title,
			coalesce(m.description, ''),
			coalesce(m.release_year, 0),
			coalesce(m.age_limit, 0),
			coalesce(m.duration_min, 0),
			m.poster_file_key,
			m.card_file_key
		from movie m
		where m.id = $1
	`

	sqlGetMovieCountriesByID = `
		select c.title
		from movie_country mc
		join country c on c.id = mc.country_id
		where mc.movie_id = $1
		order by c.title
	`

	sqlGetMovieGenresByID = `
		select g.title
		from movie_genre mg
		join genre g on g.id = mg.genre_id
		where mg.movie_id = $1
		order by g.title
	`

	sqlGetMovieActorsByID = `
		select
			a.id,
			a.name,
			a.avatar_file_key
		from movie_actor ma
		join actor a on a.id = ma.actor_id
		where ma.movie_id = $1
		order by a.name
	`

	sqlGetMovieEpisodesByID = `
		select
			e.id,
			coalesce(e.number_in_series, 0),
			coalesce(e.title, ''),
			coalesce(e.duration_sec, 0)
		from episode e
		where e.movie_id = $1
		order by e.number_in_series, e.id
	`

	sqlGetActorBaseByID = `
		select
			a.id,
			a.name,
			coalesce(a.description, ''),
			a.avatar_file_key
		from actor a
		where a.id = $1
	`

	sqlGetActorMoviesByID = `
		select
			m.id,
			m.title,
			coalesce(m.release_year, 0),
			m.poster_file_key,
			m.card_file_key
		from movie_actor ma
		join movie m on m.id = ma.movie_id
		where ma.actor_id = $1
		order by m.release_year desc, m.title
	`

	sqlGetSelectionMoviesByTitle = `
		select
			s.title,
			m.id,
			m.title,
			coalesce(m.release_year, 0),
			m.poster_file_key,
			m.card_file_key
		from selection s
		join selection_movie sm on sm.selection_id = s.id
		join movie m on m.id = sm.movie_id
		where s.title = $1
		order by sm.position, m.id
	`

	sqlGetAllSelectionMovies = `
		select
			s.title,
			m.id,
			m.title,
			coalesce(m.release_year, 0),
			m.poster_file_key,
			m.card_file_key
		from selection s
		join selection_movie sm on sm.selection_id = s.id
		join movie m on m.id = sm.movie_id
		order by s.title, sm.position, m.id
	`
)
