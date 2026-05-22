BEGIN;

DELETE FROM subscription_tariff_option
WHERE subscription_tariff_id IN (
    SELECT id FROM subscription_tariff WHERE code IN ('free', 'level_2', 'level_3', 'level_4')
)
AND subscription_option_id IN (
    SELECT id FROM subscription_option
    WHERE code IN (
        'can_watch_paid_content',
        'can_use_smart_continue',
        'ad_policy',
        'daily_coins_limit',
        'monthly_room_limit',
        'max_room_members'
    )
);

DELETE FROM subscription_tariff
WHERE code IN ('free', 'level_2', 'level_3', 'level_4');

DELETE FROM subscription_option
WHERE code IN (
    'can_watch_paid_content',
    'can_use_smart_continue',
    'ad_policy',
    'daily_coins_limit',
    'monthly_room_limit',
    'max_room_members'
);

DROP INDEX IF EXISTS subscription_tariff_option_tariff_option_unique_idx;
DROP INDEX IF EXISTS subscription_tariff_level_unique_idx;
DROP INDEX IF EXISTS subscription_tariff_code_unique_idx;

ALTER TABLE subscription_tariff
    DROP CONSTRAINT IF EXISTS subscription_tariff_payment_available_check,
    DROP CONSTRAINT IF EXISTS subscription_tariff_coins_price_required_check,
    DROP CONSTRAINT IF EXISTS subscription_tariff_money_price_required_check;

ALTER TABLE subscription_tariff
    ADD CONSTRAINT subscription_tariff_payment_available_check
        CHECK (is_coins_payment_available = true OR is_money_payment_available = true),
    ADD CONSTRAINT subscription_tariff_coins_price_required_check
        CHECK (is_coins_payment_available = false OR price_vkino_coins > 0),
    ADD CONSTRAINT subscription_tariff_money_price_required_check
        CHECK (is_money_payment_available = false OR price_money > 0);

ALTER TABLE subscription_tariff_option
    DROP COLUMN IF EXISTS value;

ALTER TABLE subscription_tariff
    DROP COLUMN IF EXISTS code,
    DROP COLUMN IF EXISTS level;

COMMIT;
