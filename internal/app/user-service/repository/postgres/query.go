package postgres

const (
	sqlGetUserRole = `
		select role from users where id = $1
	`

	sqlGetUserByEmail = `
		select
			id, email, password_hash, role, birthdate, avatar_file_key,
			registration_date, is_active, created_at, updated_at
		from users
		where email = $1
	`

	sqlGetUserByID = `
		select
			id, email, password_hash, role, birthdate, avatar_file_key,
			registration_date, is_active, created_at, updated_at
		from users
		where id = $1
	`

	sqlGetActiveSubscription = `
		select
			st.id,
			st.code,
			st.title,
			st.level,
			us.expires_at
		from user_subscription us
		join subscription_tariff st on st.id = us.subscription_tariff_id
		where us.user_id = $1
			and us.is_active = true
			and us.starts_at <= now()
			and us.expires_at > now()
			and st.is_active = true
		order by us.expires_at desc, us.id desc
		limit 1
	`

	sqlGetSubscriptionTariffByCode = `
		select
			st.id,
			st.code,
			st.title,
			st.level
		from subscription_tariff st
		where st.code = $1
			and st.is_active = true
		limit 1
	`

	sqlGetSubscriptionTariffByID = `
		select
			st.id,
			st.code,
			st.title,
			st.level,
			st.duration_days,
			st.price_vkino_coins,
			st.is_coins_payment_available
		from subscription_tariff st
		where st.id = $1
			and st.is_active = true
		limit 1
	`

	sqlDeactivateUserSubscriptions = `
		update user_subscription
		set is_active = false, updated_at = now()
		where user_id = $1
			and is_active = true
	`

	sqlCreateUserSubscription = `
		insert into user_subscription (
			user_id,
			subscription_tariff_id,
			starts_at,
			expires_at,
			is_active
		)
		values ($1, $2, $3, $4, true)
	`

	sqlGetSubscriptionTariffOptions = `
		select
			so.code,
			sto.value
		from subscription_tariff_option sto
		join subscription_option so on so.id = sto.subscription_option_id
		where sto.subscription_tariff_id = $1
			and so.is_active = true
		order by so.code
	`

	sqlGetCoinsReceivedToday = `
		select coalesce(sum(vkino_coins_count), 0)::int
		from vkino_coins_history
		where user_id = $1
			and operation_type = 'daily'
			and created_at >= date_trunc('day', now())
			and created_at < date_trunc('day', now()) + interval '1 day'
	`

	sqlGetVKinoCoinsBalance = `
		select coalesce(sum(
			case
				when operation_type in ('daily', 'signup_bonus', 'bet_win') then vkino_coins_count
				when operation_type in ('bet_lose', 'bet_place', 'purchase') then -vkino_coins_count
				else 0
			end
		), 0)::int
		from vkino_coins_history
		where user_id = $1
	`

	sqlGrantDailyVKinoCoins = `
		select grant_daily_vkino_coins($1)
	`

	sqlLockUserForUpdate = `
		select id
		from users
		where id = $1
			and is_active = true
		for update
	`

	sqlCreateCoinsPayment = `
		insert into payment (
			user_id,
			product_type,
			product_ref_id,
			amount,
			currency,
			status,
			idempotency_key,
			payment_method,
			coins_spent,
			paid_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		returning id, created_at, updated_at
	`

	sqlCreateVKinoCoinsPurchaseHistory = `
		insert into vkino_coins_history (
			user_id,
			vkino_coins_count,
			operation_type,
			description,
			operation_date
		)
		values ($1, $2, 'purchase', $3, current_date)
	`

	sqlCreateVKinoCoinsHistory = `
		insert into vkino_coins_history (
			user_id,
			vkino_coins_count,
			operation_type,
			description,
			operation_date
		)
		values ($1, $2, $3, $4, current_date)
	`

	sqlGetRoomsCreatedThisMonth = `
		select count(*)::int
		from vkino_room
		where user_creator_id = $1
			and created_at >= now() - interval '1 month'
			and created_at <= now()
	`

	sqlGetVKinoCoinsHistory = `
		with total as (
			select count(*)::int as total_count
			from vkino_coins_history
			where user_id = $1
		)
		select
			h.id,
			h.vkino_coins_count,
			h.operation_type,
			h.description,
			h.created_at,
			t.total_count
		from total t
		left join lateral (
			select
				id,
				vkino_coins_count,
				operation_type,
				description,
				created_at
			from vkino_coins_history
			where user_id = $1
			order by created_at desc, id desc
			limit $2 offset $3
		) h on true
		order by h.created_at desc nulls last, h.id desc nulls last
	`

	sqlGetFriendByID = `
		select
			u.id, u.email, u.password_hash, u.role, u.birthdate, u.avatar_file_key,
			u.registration_date, u.is_active, u.created_at, u.updated_at
		from users u
		join friend f on
			((f.user1_id = $1 and f.user2_id = u.id) or (f.user2_id = $1 and f.user1_id = u.id))
		where u.id = $2
			and u.is_active = true
	`

	sqlSearchUsersByEmail = `
		select
			u.id,
			u.email,
			coalesce(u.avatar_file_key, '') as avatar_file_key,
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
		returning
			id, email, password_hash, role, birthdate, avatar_file_key,
			registration_date, is_active, created_at, updated_at
	`

	sqlUpdateUserAvatarFileKey = `
		update users
		set avatar_file_key = $1, updated_at = now()
		where id = $2
		returning
			id, email, password_hash, role, birthdate, avatar_file_key,
			registration_date, is_active, created_at, updated_at
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

	sqlUpsertUserMovieRating = `
		insert into user_interaction (user_id, movie_id, rating)
		select $1, m.id, $3
		from movie m
		where m.id = $2
		on conflict (movie_id, user_id)
		do update set
			rating = excluded.rating,
			updated_at = now()
	`

	sqlUpsertUserMovieReview = `
		insert into user_interaction (user_id, movie_id, rating, comment)
		select $1, m.id, $3, $4
		from movie m
		where m.id = $2
		on conflict (movie_id, user_id)
		do update set
			rating = excluded.rating,
			comment = excluded.comment,
			updated_at = now()
		returning id, movie_id, rating::double precision, comment
	`

	sqlDeleteReviewReactionsByReviewID = `
		delete from user_interaction_review_reaction
		where review_id = $1
	`

	sqlGetReviewByUserAndMovie = `
		select id, is_favorite
		from user_interaction
		where user_id = $1 and movie_id = $2
	`

	sqlDeleteUserMovieReviewRow = `
		delete from user_interaction
		where id = $1
	`

	sqlClearUserMovieReview = `
		update user_interaction
		set rating = null,
			comment = null,
			updated_at = now()
		where id = $1
	`

	sqlSetReviewReaction = `
		insert into user_interaction_review_reaction (review_id, user_id, reaction)
		select ui.id, $1, $3
		from user_interaction ui
		where ui.id = $2
			and ui.user_id <> $1
			and nullif(btrim(coalesce(ui.comment, '')), '') is not null
		on conflict (review_id, user_id)
		do update set
			reaction = excluded.reaction,
			updated_at = now()
	`

	sqlGetReviewOwner = `
		select
			ui.user_id,
			nullif(btrim(coalesce(ui.comment, '')), '') is not null
		from user_interaction ui
		where ui.id = $1
	`

	sqlDeleteReviewReaction = `
		delete from user_interaction_review_reaction
		where review_id = $1 and user_id = $2
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
		with total as (
			select count(*)::int as total_count
			from user_interaction ui
			where ui.user_id = $1 and ui.is_favorite = true
		)
		select
			p.movie_id,
			t.total_count
		from total t
		left join lateral (
			select
				ui.movie_id,
				ui.updated_at
			from user_interaction ui
			where ui.user_id = $1 and ui.is_favorite = true
			order by ui.updated_at desc
			limit $2 offset $3
		) p on true
		order by p.updated_at desc nulls last
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

	sqlAcceptFriendRequestUpdate = `
		update friend_request
		set status = 'accepted'
		where id = $1 and to_user_id = $2 and status = 'pending'
		returning from_user_id, to_user_id
	`

	sqlAcceptFriendInsert = `
		insert into friend (user1_id, user2_id)
		values ($1, $2)
		on conflict do nothing
	`

	sqlAcceptFriendDeleteRequest = `
		delete from friend_request where id = $1
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
			u.email,
			coalesce(u.avatar_file_key, '') as avatar_file_key
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
