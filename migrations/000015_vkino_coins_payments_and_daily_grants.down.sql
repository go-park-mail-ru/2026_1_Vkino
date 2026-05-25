BEGIN;

DROP FUNCTION IF EXISTS grant_daily_vkino_coins(bigint);

DROP INDEX IF EXISTS vkino_coins_history_daily_once_per_day_idx;
DROP INDEX IF EXISTS idx_vkino_coins_history_user_operation_created;

ALTER TABLE vkino_coins_history
    ALTER COLUMN operation_date DROP DEFAULT;

ALTER TABLE vkino_coins_history
    DROP COLUMN IF EXISTS operation_date;

ALTER TABLE payment
    DROP CONSTRAINT IF EXISTS payment_coins_spent_check,
    DROP CONSTRAINT IF EXISTS payment_method_check;

ALTER TABLE payment
    DROP COLUMN IF EXISTS coins_spent,
    DROP COLUMN IF EXISTS payment_method;

UPDATE subscription_tariff
SET
    price_vkino_coins = 399,
    updated_at = now()
WHERE code = 'level_2';

UPDATE subscription_tariff
SET
    price_vkino_coins = 699,
    updated_at = now()
WHERE code = 'level_3';

UPDATE subscription_tariff
SET
    price_vkino_coins = 999,
    updated_at = now()
WHERE code = 'level_4';

COMMIT;
