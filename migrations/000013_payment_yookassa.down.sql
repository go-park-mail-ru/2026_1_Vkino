BEGIN;

DROP TABLE IF EXISTS payment_webhook_event;

DROP INDEX IF EXISTS payment_product_type_ref_idx;
DROP INDEX IF EXISTS payment_user_id_status_idx;

ALTER TABLE payment
    DROP CONSTRAINT IF EXISTS payment_idempotency_key_unique,
    DROP CONSTRAINT IF EXISTS payment_yookassa_payment_id_unique,
    DROP CONSTRAINT IF EXISTS payment_status_check,
    DROP CONSTRAINT IF EXISTS payment_product_type_check;

ALTER TABLE payment
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS confirmation_url,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS yookassa_payment_id,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS product_ref_id,
    DROP COLUMN IF EXISTS product_type;

ALTER TABLE payment
    ALTER COLUMN paid_at SET DEFAULT now(),
    ALTER COLUMN paid_at SET NOT NULL;

COMMIT;
