package postgres

const (

    sqlGetSupportTicketByID = `
        select id, user_id, assigned_to, category, status, support_line,
               title, description, rating, created_at, updated_at, closed_at
        from support_ticket
        where id = $1
    `

)