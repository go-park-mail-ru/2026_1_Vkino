package postgres

const (
	sqlGetUserRole = `
		select role from users where id = $1
	`

	sqlGetUserByEmail = `
		select id, email, password_hash, role, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
		from users
		where email = $1
	`

	sqlGetUserByID = `
		select id, email, password_hash, role, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
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
				where
					(f.user1_id = $1 and f.user2_id = u.id)
					or (f.user1_id = u.id and f.user2_id = $1)
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
		returning id, email, password_hash, role, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
	`

	sqlUpdateUserAvatarFileKey = `
		update users
		set avatar_file_key = $1, updated_at = now()
		where id = $2
		returning id, email, password_hash, role, birthdate, avatar_file_key, registration_date, is_active, created_at, updated_at
	`

	sqlUpsertUserFavoriteMovie = `
		insert into user_interaction (user_id, movie_id, is_favorite)
		select $1, m.id, true
		from movie m
		where m.id = $2
		on conflict (movie_id, user_id)
		do update set
			is_favorite = excluded.is_favorite
	`

	sqlToggleFavorite = `
		with current as (
			select is_favorite from user_interaction
			where user_id = $1 and movie_id = $2
		)
		insert into user_interaction (user_id, movie_id, is_favorite)
		values ($1, $2, not coalesce((select is_favorite from current), false))
		on conflict (movie_id, user_id)
		do update set
			is_favorite = not user_interaction.is_favorite
		returning is_favorite
	`

	sqlGetFavorites = `
		select m.id, m.title, m.picture_file_key
		from user_interaction ui
		join movie m on m.id = ui.movie_id
		where ui.user_id = $1 and ui.is_favorite = true
		order by ui.updated_at desc
		limit $2 offset $3
	`

	sqlCountFavorites = `
		select count(*)
		from user_interaction
		where user_id = $1 and is_favorite = true
	`

	sqlAddFriend = `
		insert into friend (user1_id, user2_id)
		values ($1, $2)
	`

	sqlDeleteFriend = `
		delete from friend
		where user1_id = $1 and user2_id = $2
	`

	sqlDeleteFriendRequestsBetweenUsers = `
		delete from friend_request
		where (from_user_id = $1 and to_user_id = $2)
			or (from_user_id = $2 and to_user_id = $1)
	`

	sqlAreFriends = `
		select exists(
			select 1
			from friend
			where user1_id = $1 and user2_id = $2
		)
	`

	sqlGetFriendRequestStatus = `
		select status
		from friend_request
		where from_user_id = $1 and to_user_id = $2
	`

	sqlDeleteFriendRequestPair = `
		delete from friend_request
		where from_user_id = $1 and to_user_id = $2
	`

	sqlSendFriendRequest = `
		insert into friend_request (from_user_id, to_user_id, status)
		values ($1, $2, 'pending')
		on conflict (from_user_id, to_user_id)
		do update set
			status = case when friend_request.status = 'declined' then 'pending' else friend_request.status end
		returning id
	`

	sqlRespondToRequest = `
		update friend_request
		set status = $1
		where id = $2 and to_user_id = $3 and status = 'pending'
		returning from_user_id
	`

	sqlDeleteOutgoingRequest = `
		delete from friend_request
		where id = $1 and from_user_id = $2 and status = 'pending'
		returning to_user_id
	`

	sqlAcceptFriendRequestAtomic = `
		with updated as (
			update friend_request
			set status = 'accepted'
			where id = $1 and to_user_id = $2 and status = 'pending'
			returning from_user_id, to_user_id, id
		),
		inserted as (
			insert into friend (user1_id, user2_id)
			select
				case
					when from_user_id <= to_user_id then from_user_id
					else to_user_id
				end,
				case
					when from_user_id >= to_user_id then from_user_id
					else to_user_id
				end
			from updated
			on conflict do nothing
		),
		deleted as (
			delete from friend_request fr
			using updated u
			where fr.id = u.id
		)
		select from_user_id from updated
	`

	sqlGetIncomingRequests = `
		select fr.id, fr.from_user_id, u.email, fr.created_at
		from friend_request fr
		join users u on u.id = fr.from_user_id
		where fr.to_user_id = $1 and fr.status = 'pending'
		order by fr.created_at desc
		limit $2
	`

	sqlGetOutgoingRequests = `
		select fr.id, fr.to_user_id, u.email, fr.created_at
		from friend_request fr
		join users u on u.id = fr.to_user_id
		where fr.from_user_id = $1 and fr.status = 'pending'
		order by fr.created_at desc
		limit $2
	`

	sqlGetFriends = `
		select
			case when f.user1_id = $1 then f.user2_id else f.user1_id end as friend_id,
			u.email
		from friend f
		join users u on u.id = case when f.user1_id = $1 then f.user2_id else f.user1_id end
		where (f.user1_id = $1 or f.user2_id = $1) and u.is_active = true
		order by u.email
		limit $2 offset $3
	`

	sqlCountFriends = `
		select count(*)
		from friend f
		join users u on u.id = case when f.user1_id = $1 then f.user2_id else f.user1_id end
		where (f.user1_id = $1 or f.user2_id = $1) and u.is_active = true
	`
)