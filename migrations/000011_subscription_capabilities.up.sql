BEGIN;

ALTER TABLE subscription_tariff
    ADD COLUMN IF NOT EXISTS code text,
    ADD COLUMN IF NOT EXISTS level int;

ALTER TABLE subscription_tariff_option
    ADD COLUMN IF NOT EXISTS value text;

UPDATE subscription_tariff
SET code = 'legacy_' || id
WHERE code IS NULL OR char_length(trim(code)) = 0;

UPDATE subscription_tariff
SET level = 1000 + id
WHERE level IS NULL OR level < 1;

ALTER TABLE subscription_tariff
    DROP CONSTRAINT IF EXISTS subscription_tariff_payment_available_check,
    DROP CONSTRAINT IF EXISTS subscription_tariff_coins_price_required_check,
    DROP CONSTRAINT IF EXISTS subscription_tariff_money_price_required_check;

ALTER TABLE subscription_tariff
    ALTER COLUMN code SET NOT NULL,
    ALTER COLUMN level SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS subscription_tariff_code_unique_idx
    ON subscription_tariff (code);

CREATE UNIQUE INDEX IF NOT EXISTS subscription_tariff_level_unique_idx
    ON subscription_tariff (level);

CREATE UNIQUE INDEX IF NOT EXISTS subscription_tariff_option_tariff_option_unique_idx
    ON subscription_tariff_option (subscription_tariff_id, subscription_option_id);

ALTER TABLE subscription_tariff
    ADD CONSTRAINT subscription_tariff_payment_available_check
        CHECK (
            code = 'free'
            OR is_coins_payment_available = true
            OR is_money_payment_available = true
        ),
    ADD CONSTRAINT subscription_tariff_coins_price_required_check
        CHECK (
            code = 'free'
            OR is_coins_payment_available = false
            OR price_vkino_coins > 0
        ),
    ADD CONSTRAINT subscription_tariff_money_price_required_check
        CHECK (
            code = 'free'
            OR is_money_payment_available = false
            OR price_money > 0
        );

INSERT INTO subscription_option (code, title, description, is_active)
VALUES
    ('can_watch_paid_content', 'Платный контент', 'Доступ к платным фильмам и сериям.', true),
    ('can_use_smart_continue', 'Умное продолжение', 'Доступ к разделу умного продолжения просмотра.', true),
    ('ad_policy', 'Режим рекламы', 'Правила показа и пропуска рекламы.', true),
    ('daily_coins_limit', 'Дневной лимит монет', 'Сколько VKino coins можно получить за день.', true),
    ('monthly_room_limit', 'Лимит комнат в месяц', 'Сколько комнат совместного просмотра можно создать за месяц.', true),
    ('max_room_members', 'Лимит активных участников комнаты', 'Максимум активных пользователей в комнате.', true)
ON CONFLICT (code) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = now();

INSERT INTO subscription_tariff (
    code,
    level,
    title,
    description,
    price_vkino_coins,
    is_coins_payment_available,
    price_money,
    is_money_payment_available,
    duration_days,
    is_active
)
VALUES
    ('free', 1, 'Free', 'Базовый тариф с бесплатным контентом и минимальными лимитами.', 0, false, 0, false, 365, true),
    ('level_2', 2, 'Level 2', 'Подписка с доступом к платному контенту и базовым smart-функциям.', 399, true, 299, true, 30, true),
    ('level_3', 3, 'Level 3', 'Продвинутый тариф с полным пропуском рекламы и расширенными лимитами.', 699, true, 499, true, 30, true),
    ('level_4', 4, 'Level 4', 'Максимальный тариф без месячного лимита комнат и без рекламы.', 999, true, 699, true, 30, true)
ON CONFLICT (code) DO UPDATE
SET level = EXCLUDED.level,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    price_vkino_coins = EXCLUDED.price_vkino_coins,
    is_coins_payment_available = EXCLUDED.is_coins_payment_available,
    price_money = EXCLUDED.price_money,
    is_money_payment_available = EXCLUDED.is_money_payment_available,
    duration_days = EXCLUDED.duration_days,
    is_active = EXCLUDED.is_active,
    updated_at = now();

WITH tariff_options AS (
    SELECT *
    FROM (
        VALUES
            ('free', 'can_watch_paid_content', 'false'),
            ('free', 'can_use_smart_continue', 'false'),
            ('free', 'ad_policy', 'no_skip'),
            ('free', 'daily_coins_limit', '3'),
            ('free', 'monthly_room_limit', '3'),
            ('free', 'max_room_members', '2'),
            ('level_2', 'can_watch_paid_content', 'true'),
            ('level_2', 'can_use_smart_continue', 'true'),
            ('level_2', 'ad_policy', 'skip_preroll'),
            ('level_2', 'daily_coins_limit', '6'),
            ('level_2', 'monthly_room_limit', '6'),
            ('level_2', 'max_room_members', '2'),
            ('level_3', 'can_watch_paid_content', 'true'),
            ('level_3', 'can_use_smart_continue', 'true'),
            ('level_3', 'ad_policy', 'skip_all'),
            ('level_3', 'daily_coins_limit', '12'),
            ('level_3', 'monthly_room_limit', '10'),
            ('level_3', 'max_room_members', '4'),
            ('level_4', 'can_watch_paid_content', 'true'),
            ('level_4', 'can_use_smart_continue', 'true'),
            ('level_4', 'ad_policy', 'none'),
            ('level_4', 'daily_coins_limit', '30'),
            ('level_4', 'monthly_room_limit', NULL),
            ('level_4', 'max_room_members', '4')
    ) AS seeded(tariff_code, option_code, option_value)
)
INSERT INTO subscription_tariff_option (subscription_tariff_id, subscription_option_id, value)
SELECT
    st.id,
    so.id,
    seeded.option_value
FROM tariff_options seeded
JOIN subscription_tariff st ON st.code = seeded.tariff_code
JOIN subscription_option so ON so.code = seeded.option_code
ON CONFLICT (subscription_tariff_id, subscription_option_id) DO UPDATE
SET value = EXCLUDED.value;

COMMIT;
