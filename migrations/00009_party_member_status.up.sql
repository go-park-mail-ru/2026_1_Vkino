BEGIN;

alter table vkino_room_member
    add column if not exists status text not null default 'active';

update vkino_room_member
set status = 'active'
where status not in ('pending', 'active');

alter table vkino_room_member
    drop constraint if exists vkino_room_member_status_check;

alter table vkino_room_member
    add constraint vkino_room_member_status_check
        check (status in ('pending', 'active'));

COMMIT;
