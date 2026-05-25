BEGIN;

UPDATE subscription_tariff
SET
    price_vkino_coins = 20,
    is_coins_payment_available = true,
    updated_at = now()
WHERE code = 'level_2';

UPDATE subscription_tariff
SET
    price_vkino_coins = 50,
    is_coins_payment_available = true,
    updated_at = now()
WHERE code = 'level_3';

UPDATE subscription_tariff
SET
    price_vkino_coins = 150,
    is_coins_payment_available = true,
    updated_at = now()
WHERE code = 'level_4';

ALTER TABLE payment
    ADD COLUMN IF NOT EXISTS payment_method text NOT NULL DEFAULT 'yookassa',
    ADD COLUMN IF NOT EXISTS coins_spent int;

ALTER TABLE payment
    DROP CONSTRAINT IF EXISTS payment_method_check,
    DROP CONSTRAINT IF EXISTS payment_coins_spent_check;

ALTER TABLE payment
    ADD CONSTRAINT payment_method_check
        CHECK (payment_method IN ('yookassa', 'vkino_coins')),
    ADD CONSTRAINT payment_coins_spent_check
        CHECK (coins_spent IS NULL OR coins_spent > 0);

CREATE INDEX IF NOT EXISTS idx_vkino_coins_history_user_operation_created
    ON vkino_coins_history (user_id, operation_type, created_at DESC, id DESC);

ALTER TABLE vkino_coins_history
    ADD COLUMN IF NOT EXISTS operation_date date;

UPDATE vkino_coins_history
SET operation_date = created_at::date
WHERE operation_date IS NULL;

ALTER TABLE vkino_coins_history
    ALTER COLUMN operation_date SET DEFAULT CURRENT_DATE,
    ALTER COLUMN operation_date SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS vkino_coins_history_daily_once_per_day_idx
    ON vkino_coins_history (user_id, operation_type, operation_date)
    WHERE operation_type = 'daily';

CREATE OR REPLACE FUNCTION grant_daily_vkino_coins(p_user_id bigint)
RETURNS int
LANGUAGE plpgsql
AS $$
DECLARE
    v_tariff_id bigint;
    v_daily_coins int := 0;
    v_granted int := 0;
BEGIN
    SELECT st.id
    INTO v_tariff_id
    FROM user_subscription us
    JOIN subscription_tariff st ON st.id = us.subscription_tariff_id
    WHERE us.user_id = p_user_id
        AND us.is_active = true
        AND us.starts_at <= now()
        AND us.expires_at > now()
        AND st.is_active = true
    ORDER BY us.expires_at DESC, us.id DESC
    LIMIT 1;

    IF v_tariff_id IS NULL THEN
        SELECT st.id
        INTO v_tariff_id
        FROM subscription_tariff st
        WHERE st.code = 'free'
            AND st.is_active = true
        LIMIT 1;
    END IF;

    IF v_tariff_id IS NULL THEN
        RETURN 0;
    END IF;

    SELECT COALESCE(NULLIF(BTRIM(sto.value), '')::int, 0)
    INTO v_daily_coins
    FROM subscription_tariff_option sto
    JOIN subscription_option so ON so.id = sto.subscription_option_id
    WHERE sto.subscription_tariff_id = v_tariff_id
        AND so.code = 'daily_coins_limit'
        AND so.is_active = true
    LIMIT 1;

    IF COALESCE(v_daily_coins, 0) <= 0 THEN
        RETURN 0;
    END IF;

    INSERT INTO vkino_coins_history (
        user_id,
        vkino_coins_count,
        operation_type,
        description,
        operation_date
    )
    VALUES (
        p_user_id,
        v_daily_coins,
        'daily',
        'Ежедневное начисление VKino coins',
        CURRENT_DATE
    )
    ON CONFLICT (user_id, operation_type, operation_date)
        WHERE operation_type = 'daily'
        DO NOTHING
    RETURNING vkino_coins_count INTO v_granted;

    RETURN COALESCE(v_granted, 0);
END;
$$;

COMMIT;
