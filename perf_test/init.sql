\set ON_ERROR_STOP on

-- Base schema before perf optimizations.
-- Intentionally excludes:
--   - 00002_migration.fill.up.sql (seed data, not DDL)
--   - 00008_perf_optimizations.up.sql (the optimization migration itself)

\ir ../migrations/00001_migration.up.sql
\ir ../migrations/00003_migration.search.up.sql
\ir ../migrations/00004_migration.support.up.sql
\ir ../migrations/00005_friend_requests.up.sql
\ir ../migrations/00006_migration.coins.up.sql
\ir ../migrations/00007_movie_external_ratings.up.sql
