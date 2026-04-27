package postgres

const (
	sqlCreateSupportTicket = `
		insert into support_ticket (
			user_id,
			user_email,
			support_line,
			category,
			title,
			description,
			attachment_file_key
		)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning
			id,
			user_id,
			user_email,
			category,
			status,
			support_line,
			title,
			description,
			attachment_file_key,
			rating,
			created_at,
			updated_at,
			closed_at
	`

	sqlGetSupportTicketByID = `
		select
			id,
			user_id,
			user_email,
			category,
			status,
			support_line,
			title,
			description,
			attachment_file_key,
			rating,
			created_at,
			updated_at,
			closed_at
		from support_ticket
		where id = $1
	`

	sqlGetSupportTickets = `
		select
			id,
			user_id,
			user_email,
			category,
			status,
			support_line,
			title,
			description,
			attachment_file_key,
			rating,
			created_at,
			updated_at,
			closed_at
		from support_ticket
		where
			($1 = 0 or user_id = $1)
			and ($2 = '' or user_email = $2)
			and ($3 = '' or status = $3)
			and ($4 = '' or category = $4)
			and ($5 = 0 or support_line = $5)
			and (coalesce(array_length($6::text[], 1), 0) = 0 or category = any($6::text[]))
		order by created_at desc
	`

	sqlUpdateSupportTicket = `
		update support_ticket
		set
			category = coalesce(nullif($2, ''), category),
			status = coalesce(nullif($3, ''), status),
			support_line = case
				when $4 = 0 then support_line
				else $4
			end,
			title = coalesce(nullif($5, ''), title),
			description = coalesce(nullif($6, ''), description),
			attachment_file_key = coalesce(nullif($7, ''), attachment_file_key),
			user_email = coalesce(nullif($8, ''), user_email),
			closed_at = case
				when $3 in ('resolved', 'closed') then now()
				else closed_at
			end
		where id = $1
		returning
			id,
			user_id,
			user_email,
			category,
			status,
			support_line,
			title,
			description,
			attachment_file_key,
			rating,
			created_at,
			updated_at,
			closed_at
	`

	sqlDeleteSupportTicket = `
		delete from support_ticket
		where id = $1
	`

	sqlCreateSupportTicketMessage = `
		insert into support_ticket_message (
			ticket_id,
			sender_id,
			content,
			content_file_key
		)
		values ($1, $2, $3, $4)
		returning
			id,
			ticket_id,
			sender_id,
			content,
			content_file_key,
			created_at
	`

	sqlGetSupportTicketMessages = `
		select
			id,
			ticket_id,
			sender_id,
			content,
			content_file_key,
			created_at
		from support_ticket_message
		where ticket_id = $1
		order by created_at asc
	`

	sqlRateSupportTicket = `
		update support_ticket
		set rating = $2
		where id = $1
			and user_id = $3
		and status in ('resolved', 'closed')
		returning
			id,
			user_id,
			user_email,
			category,
			status,
			support_line,
			title,
			description,
			attachment_file_key,
			rating,
			created_at,
			updated_at,
			closed_at
	`

	sqlGetSupportStatistics = `
		select
			count(*) as total,
			count(*) filter (where status = 'open') as open,
			count(*) filter (where status = 'in_progress') as in_progress,
			count(*) filter (where status = 'waiting_user') as waiting_user,
			count(*) filter (where status = 'resolved') as resolved,
			count(*) filter (where status = 'closed') as closed,
			coalesce(avg(rating), 0) as average_rating
		from support_ticket
		where coalesce(array_length($1::text[], 1), 0) = 0 or category = any($1::text[])
	`

	sqlGetSupportStatisticsByCategory = `
		select
			category,
			count(*)
		from support_ticket
		group by category
		order by category
	`

	sqlGetSupportStatisticsByLine = `
		select
			support_line,
			count(*)
		from support_ticket
		group by support_line
		order by support_line
	`
)
