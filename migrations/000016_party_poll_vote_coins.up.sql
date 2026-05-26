BEGIN;

ALTER TABLE vkino_room_chat_bet_answer
    ADD COLUMN IF NOT EXISTS vkino_coins_count int NOT NULL DEFAULT 0;

ALTER TABLE vkino_room_chat_bet_answer
    DROP CONSTRAINT IF EXISTS vkino_room_chat_bet_answer_vkino_coins_count_check;

ALTER TABLE vkino_room_chat_bet_answer
    ADD CONSTRAINT vkino_room_chat_bet_answer_vkino_coins_count_check
        CHECK (vkino_coins_count >= 0);

ALTER TABLE vkino_coins_history
    DROP CONSTRAINT IF EXISTS vkino_coins_history_operation_type_check;

ALTER TABLE vkino_coins_history
    ADD CONSTRAINT vkino_coins_history_operation_type_check
        CHECK (operation_type IN ('daily', 'bet_win', 'bet_lose', 'bet_place', 'purchase'));

COMMIT;
