#!/bin/sh

set -eu

export PGPASSWORD="${POSTGRES_ADMIN_PASSWORD}"

psql \
  -v ON_ERROR_STOP=1 \
  -h db \
  -U "${POSTGRES_ADMIN_USER}" \
  -d "${POSTGRES_DB}" \
  -f /scripts/02_extensions.sql

psql \
  -v ON_ERROR_STOP=1 \
  -h db \
  -U "${POSTGRES_ADMIN_USER}" \
  -d "${POSTGRES_DB}" \
  -v db_name="${POSTGRES_DB}" \
  -v auth_user="${VKINO_AUTH_DB_USER}" \
  -v auth_password="${VKINO_AUTH_DB_PASSWORD}" \
  -v user_user="${VKINO_USER_DB_USER}" \
  -v user_password="${VKINO_USER_DB_PASSWORD}" \
  -v movie_user="${VKINO_MOVIE_DB_USER}" \
  -v movie_password="${VKINO_MOVIE_DB_PASSWORD}" \
  -v migrator_user="${VKINO_MIGRATOR_DB_USER}" \
  -v migrator_password="${VKINO_MIGRATOR_DB_PASSWORD}" \
  -v monitoring_user="${VKINO_MONITORING_DB_USER}" \
  -v monitoring_password="${VKINO_MONITORING_DB_PASSWORD}" \
  -f /scripts/01_admin_setup.sql
