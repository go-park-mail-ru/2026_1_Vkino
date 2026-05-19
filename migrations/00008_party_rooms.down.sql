BEGIN;

drop table if exists vkino_room_playback_state;
drop table if exists vkino_room_invite;
drop table if exists vkino_room_member;

alter table if exists vkino_room
    drop column if exists visibility,
    drop column if exists name;

COMMIT;
