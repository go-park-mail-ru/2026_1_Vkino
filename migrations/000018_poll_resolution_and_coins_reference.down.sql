BEGIN;

DROP INDEX IF EXISTS vkino_coins_history_reference_key_unique_idx;

ALTER TABLE vkino_coins_history
    DROP CONSTRAINT IF EXISTS vkino_coins_history_reference_key_length,
    DROP COLUMN IF EXISTS reference_key;

ALTER TABLE vkino_room_chat_bet
    DROP COLUMN IF EXISTS resolved_at,
    DROP COLUMN IF EXISTS resolved_by_user_id,
    DROP COLUMN IF EXISTS resolved_bet_variant_id;

COMMIT;
