BEGIN;

alter table friend_request
    add constraint friend_request_status_check
        check (status in ('pending', 'accepted', 'declined'));

COMMIT;
