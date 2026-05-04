package postgres

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
		select refresh_token, expires_at
		from user_session
		where user_id = $1
	`

	sqlDeleteSession = `
		delete from user_session
		where user_id = $1
	`

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

	sqlCreateUser = `
		insert into users (email, password_hash)
		values ($1, $2)
		returning id, email, password_hash, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
	`

	//nolint:gosec // References the schema column name, not a hardcoded credential.
	sqlUpdateUserPasswordByID = `
		update users
		set password_hash = $1, updated_at = now()
		where id = $2
	`
)
