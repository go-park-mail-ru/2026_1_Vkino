BEGIN;

alter table friend_request
    drop constraint if exists friend_request_status_check;

COMMIT;
