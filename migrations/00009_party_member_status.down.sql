BEGIN;

alter table vkino_room_member
    drop constraint if exists vkino_room_member_status_check;

alter table vkino_room_member
    drop column if exists status;

COMMIT;
