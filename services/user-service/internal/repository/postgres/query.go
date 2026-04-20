package postgres

const (
	sqlGetUserByEmail = `
		select id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
		from users
		where email = $1
	`

	sqlGetUserByID = `
		select id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
		from users
		where id = $1
	`

	sqlSearchUsersByEmail = `
		select
			u.id,
			u.email,
			exists(
				select 1
				from friend f
				where f.user1_id = least($1, u.id) and f.user2_id = greatest($1, u.id)
			) as is_friend
		from users u
		where u.id <> $1
			and u.is_active = true
			and u.email ilike '%' || $2 || '%'
		order by u.email
		limit 20
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

	sqlUpsertUserFavoriteMovie = `
		insert into user_interaction (user_id, movie_id, is_favorite)
		select $1, m.id, true
		from movie m
		where m.id = $2
		on conflict (movie_id, user_id)
		do update set
			is_favorite = excluded.is_favorite,
			updated_at = now()
	`

	sqlAddFriend = `
		insert into friend (user1_id, user2_id)
		values (least($1, $2), greatest($1, $2))
	`

	sqlDeleteFriend = `
		delete from friend
		where user1_id = least($1, $2) and user2_id = greatest($1, $2)
	`
)
