BEGIN;

-- Safe no-op: these links may already exist before this migration,
-- so deleting them here could remove valid data not created by this file.

COMMIT;