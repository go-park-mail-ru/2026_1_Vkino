BEGIN;

create index if not exists idx_vkino_coins_history_user_created_id
    on vkino_coins_history (user_id, created_at desc, id desc);

COMMIT;
