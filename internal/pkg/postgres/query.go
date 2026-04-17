package postgres

// Movie queries
const (
	sqlGetSelectionByTitle = `
		select m.id, m.title, m.picture_file_key
		from movie m 
		join movie_to_selection mts ON (mts.movie_id = m.id)
		where mts.selection_id = (select id from selection where title=$1)
	`

	sqlGetAllSelectionTitles = `
		select title from selection
	`

	sqlGetMovieByID = `
		select id, title, description, director, content_type, release_year, 
		duration_seconds, age_limit, original_language_id, country_id, picture_file_key, coalesce(poster_file_key, '')
		from movie where id=$1
	`

	sqlGetGenresByMovieID = `
		select g.title
		from genre_to_movie gtm
		join genre g on g.id = gtm.genre_id
		where gtm.movie_id = $1
		order by g.title
	`

	sqlGetActorsByMovieID = `
		select a.id, a.full_name, a.picture_file_key
		from actor_to_movie atm
		join actor a on a.id = atm.actor_id
		where atm.movie_id = $1
		order by a.full_name
	`

	sqlGetActorByID = `
		select
			id,
			full_name,
			coalesce(to_char(birthdate, 'YYYY-MM-DD'), ''),
			coalesce(biography, ''),
			country_id,
			picture_file_key
		from actor 
		where id = $1
	`

	sqlGetMoviesByActorID = `
		select m.id, m.title, m.picture_file_key
		from actor_to_movie atm
		join movie m on m.id = atm.movie_id
		where atm.actor_id = $1
		order by m.title
	`

	sqlGetEpisodesByMovieID = `
		select
			id,
			movie_id,
			season_number,
			episode_number,
			coalesce(title, ''),
			coalesce(description, ''),
			duration_seconds,
			picture_file_key
		from episode
		where movie_id = $1
		order by season_number, episode_number
	`

	sqlGetEpisodePlayback = `
		select
			id,
			movie_id,
			season_number,
			episode_number,
			coalesce(title, ''),
			duration_seconds,
			video_file_key
		from episode
		where id = $1
	`

	sqlGetWatchProgress = `
		select position_seconds
		from watch_progress_episode
		where user_id = $1 and episode_id = $2
	`

	sqlUpsertWatchProgress = `
		insert into watch_progress_episode (user_id, episode_id, position_seconds)
		values ($1, $2, $3)
		on conflict (user_id, episode_id)
		do update set
			position_seconds = excluded.position_seconds,
			updated_at = now()
	`
)

// Session queries
const (
	sqlSaveSession = `
		insert into user_session (user_id, refresh_token, expires_at) 
		values ($1, $2, $3)
		on conflict (user_id) 
		do update set 
		refresh_token = excluded.refresh_token,
		expires_at = excluded.expires_at
	`

	sqlGetSession = `
		select refresh_token, expires_at from user_session where user_id=$1
	`

	sqlDeleteSession = `
		delete from user_session where user_id=$1
	`
)

// User queries
const (
	sqlGetUserByEmail = `
		select id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at 
		from users where email = $1
	`

	sqlGetUserByID = `
		select id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at 
		from users where id = $1
	`

	sqlCreateUser = `insert into users (email, password_hash) 
		values ($1, $2)
		returning id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
	`

	sqlUpdateUser = `
		update users 
		set password_hash = $1, updated_at = $2 
		where email = $3 
		returning id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
	`

	sqlUpdateUserBirthdate = `
		update users
		set birthdate = $1, updated_at = now()
		where id = $2
		returning id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
	`

	sqlUpdateUserAvatarFileKey = `
		update users
		set avatar_file_key = $1, updated_at = now()
		where id = $2
		returning id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
	`

	sqlUpdateUserPasswordByID = `
		update users
		set password_hash = $1, updated_at = now()
		where id = $2
	`

	sqlDeleteUser = `
		delete from users where email = $1
	`
)
