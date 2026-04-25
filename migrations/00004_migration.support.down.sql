BEGIN;

drop table if exists support_ticket_message;
drop table if exists support_ticket;

alter table users
drop column if exists role;

COMMIT;