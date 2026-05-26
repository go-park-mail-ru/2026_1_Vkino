BEGIN;

ALTER TABLE vkino_room_chat_bet
    ADD COLUMN IF NOT EXISTS resolved_bet_variant_id bigint NULL
        REFERENCES vkino_room_chat_bet_variant (id)
            ON UPDATE CASCADE
            ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS resolved_by_user_id bigint NULL
        REFERENCES users (id)
            ON UPDATE CASCADE
            ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS resolved_at timestamptz NULL;

ALTER TABLE vkino_coins_history
    ADD COLUMN IF NOT EXISTS reference_key text NULL
        CONSTRAINT vkino_coins_history_reference_key_length
            CHECK (reference_key IS NULL OR (char_length(reference_key) > 0 AND char_length(reference_key) <= 255));

CREATE UNIQUE INDEX IF NOT EXISTS vkino_coins_history_reference_key_unique_idx
    ON vkino_coins_history (reference_key)
    WHERE reference_key IS NOT NULL;

COMMIT;
