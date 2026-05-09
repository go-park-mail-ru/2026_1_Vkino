//nolint:lll // SQL definitions are kept close to scan order for readability.
package postgres

const (
	sqlGetOverviewActiveRooms = `
		select
			r.id,
			r.name,
			r.visibility,
			coalesce(inv.invite_code, ''),
			r.user_owner_id,
			owner.email,
			coalesce(member_counts.participants_count, 0),
			coalesce(ps.movie_id, 0),
			coalesce(ps.episode_id, 0),
			coalesce(ps.playback_url, ''),
			coalesce(ps.duration_seconds, 0),
			coalesce(ps.position_seconds, 0),
			coalesce(ps.status, 'paused'),
			coalesce(ps.updated_at, r.updated_at),
			r.updated_at
		from vkino_room r
		join users owner on owner.id = r.user_owner_id
		left join (
			select vkino_room_id, count(*)::int as participants_count
			from vkino_room_member
			group by vkino_room_id
		) member_counts on member_counts.vkino_room_id = r.id
		left join vkino_room_invite inv on inv.vkino_room_id = r.id
		left join vkino_room_playback_state ps on ps.vkino_room_id = r.id
		order by r.updated_at desc, r.id desc
		limit 20
	`

	sqlGetOverviewMyRooms = `
		select
			r.id,
			r.name,
			r.visibility,
			coalesce(inv.invite_code, ''),
			r.user_owner_id,
			owner.email,
			coalesce(member_counts.participants_count, 0),
			coalesce(ps.movie_id, 0),
			coalesce(ps.episode_id, 0),
			coalesce(ps.playback_url, ''),
			coalesce(ps.duration_seconds, 0),
			coalesce(ps.position_seconds, 0),
			coalesce(ps.status, 'paused'),
			coalesce(ps.updated_at, r.updated_at),
			r.updated_at
		from vkino_room_member rm
		join vkino_room r on r.id = rm.vkino_room_id
		join users owner on owner.id = r.user_owner_id
		left join (
			select vkino_room_id, count(*)::int as participants_count
			from vkino_room_member
			group by vkino_room_id
		) member_counts on member_counts.vkino_room_id = r.id
		left join vkino_room_invite inv on inv.vkino_room_id = r.id
		left join vkino_room_playback_state ps on ps.vkino_room_id = r.id
		where rm.user_id = $1
		order by r.updated_at desc, r.id desc
	`

	sqlGetRoomBaseByID = `
		select
			r.id,
			r.name,
			r.visibility,
			r.user_owner_id,
			coalesce(inv.invite_code, ''),
			r.updated_at
		from vkino_room r
		left join vkino_room_invite inv on inv.vkino_room_id = r.id
		where r.id = $1
	`

	sqlGetRoomMembers = `
		select
			rm.user_id,
			u.email,
			coalesce(u.avatar_file_key, ''),
			rm.role,
			rm.created_at
		from vkino_room_member rm
		join users u on u.id = rm.user_id
		where rm.vkino_room_id = $1
		order by
			case when rm.role = 'host' then 0 else 1 end,
			rm.created_at,
			rm.user_id
	`

	sqlGetRoomPlaybackState = `
		select
			coalesce(movie_id, 0),
			coalesce(episode_id, 0),
			playback_url,
			duration_seconds,
			position_seconds,
			status,
			updated_at
		from vkino_room_playback_state
		where vkino_room_id = $1
	`

	sqlGetRoomMessages = `
		select
			m.id,
			m.vkino_room_id,
			m.user_id,
			u.email,
			coalesce(m.content, ''),
			m.created_at
		from vkino_room_chat_message m
		join users u on u.id = m.user_id
		where m.vkino_room_id = $1
		order by m.created_at asc, m.id asc
		limit 100
	`

	sqlGetRoomPolls = `
		select
			b.id,
			b.vkino_room_id,
			b.bet_title,
			b.user_creator_id,
			b.created_at,
			v.id,
			v.bet_variant,
			coalesce(vote_counts.votes_count, 0)
		from vkino_room_chat_bet b
		left join vkino_room_chat_bet_variant v on v.vkino_room_chat_bet_id = b.id
		left join (
			select bet_variant_id, count(*)::bigint as votes_count
			from vkino_room_chat_bet_answer
			group by bet_variant_id
		) vote_counts on vote_counts.bet_variant_id = v.id
		where b.vkino_room_id = $1
		order by b.created_at asc, b.id asc, v.id asc
	`

	sqlCreateRoom = `
		insert into vkino_room (user_creator_id, user_owner_id, name, visibility)
		values ($1, $1, $2, $3)
		returning id, updated_at
	`

	sqlCreateRoomMember = `
		insert into vkino_room_member (vkino_room_id, user_id, role)
		values ($1, $2, $3)
		on conflict (vkino_room_id, user_id)
		do update set role = excluded.role, updated_at = now()
	`

	sqlCreateRoomInvite = `
		insert into vkino_room_invite (vkino_room_id, created_by_user_id, invite_code)
		values ($1, $2, $3)
		returning invite_code
	`

	sqlCreateRoomPlaybackState = `
		insert into vkino_room_playback_state (
			vkino_room_id,
			movie_id,
			episode_id,
			playback_url,
			duration_seconds,
			position_seconds,
			status
		) values ($1, nullif($2, 0), nullif($3, 0), '', 0, 0, 'paused')
		on conflict (vkino_room_id)
		do nothing
	`

	sqlAddRoomMember = `
		insert into vkino_room_member (vkino_room_id, user_id, role)
		values ($1, $2, 'member')
		on conflict (vkino_room_id, user_id)
		do nothing
	`

	sqlDeleteRoom = `
		delete from vkino_room
		where id = $1
	`

	sqlGetInviteByCode = `
		select
			vkino_room_id,
			invite_code,
			created_by_user_id,
			created_at,
			expires_at
		from vkino_room_invite
		where invite_code = $1
	`

	sqlUpsertPlaybackState = `
		insert into vkino_room_playback_state (
			vkino_room_id,
			movie_id,
			episode_id,
			playback_url,
			duration_seconds,
			position_seconds,
			status
		) values ($1, nullif($2, 0), nullif($3, 0), $4, $5, $6, $7)
		on conflict (vkino_room_id)
		do update set
			movie_id = excluded.movie_id,
			episode_id = excluded.episode_id,
			playback_url = excluded.playback_url,
			duration_seconds = excluded.duration_seconds,
			position_seconds = excluded.position_seconds,
			status = excluded.status,
			updated_at = now()
	`

	sqlInsertRoomMessage = `
		insert into vkino_room_chat_message (user_id, vkino_room_id, content)
		values ($1, $2, $3)
		returning id, created_at
	`

	sqlInsertRoomPoll = `
		insert into vkino_room_chat_bet (user_creator_id, vkino_room_id, vkino_coins_count, bet_title)
		values ($1, $2, $3, $4)
		returning id, created_at
	`

	sqlInsertRoomPollOption = `
		insert into vkino_room_chat_bet_variant (vkino_room_chat_bet_id, bet_variant)
		values ($1, $2)
		returning id
	`

	sqlInsertRoomVote = `
		insert into vkino_room_chat_bet_answer (user_id, bet_variant_id)
		values ($1, $2)
		on conflict (user_id, bet_variant_id)
		do nothing
	`
)
