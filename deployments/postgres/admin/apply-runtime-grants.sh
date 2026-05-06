#!/bin/sh

set -eu

export PGPASSWORD="${POSTGRES_ADMIN_PASSWORD}"

psql \
  -v ON_ERROR_STOP=1 \
  -h db \
  -U "${POSTGRES_ADMIN_USER}" \
  -d "${POSTGRES_DB}" \
  -v auth_user="${VKINO_AUTH_DB_USER}" \
  -v user_user="${VKINO_USER_DB_USER}" \
  -v movie_user="${VKINO_MOVIE_DB_USER}" \
  -v migrator_user="${VKINO_MIGRATOR_DB_USER}" \
  -v monitoring_user="${VKINO_MONITORING_DB_USER}" \
  -f /scripts/03_runtime_grants.sql
