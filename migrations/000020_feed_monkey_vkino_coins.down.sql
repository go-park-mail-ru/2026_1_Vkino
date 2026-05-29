ALTER TABLE vkino_coins_history
    DROP CONSTRAINT IF EXISTS vkino_coins_history_operation_type_check;

ALTER TABLE vkino_coins_history
    ADD CONSTRAINT vkino_coins_history_operation_type_check
        CHECK (operation_type IN (
            'daily',
            'signup_bonus',
            'bet_win',
            'bet_lose',
            'bet_place',
            'purchase',
            'coins_purchase'
        ));
