#!/usr/bin/env bash

set -Eeuo pipefail

DEPLOY_ENV="${1:-dev}"
MOVIE_COUNT="${MOVIE_COUNT:-100000}"
USER_COUNT="${USER_COUNT:-20000}"
ACTOR_COUNT="${ACTOR_COUNT:-20000}"
SELECTION_COUNT="${SELECTION_COUNT:-40}"
MOVIES_PER_SELECTION="${MOVIES_PER_SELECTION:-400}"
INTERACTIONS_PER_MOVIE="${INTERACTIONS_PER_MOVIE:-12}"
WATCH_ROWS_PER_USER="${WATCH_ROWS_PER_USER:-12}"

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/deployments/${DEPLOY_ENV}/compose.yaml"
ENV_FILE="${ROOT_DIR}/deployments/${DEPLOY_ENV}/.env"
GENERATE_SQL="${ROOT_DIR}/perf_test/sql/generate_test_data.sql"
CLEANUP_SQL="${ROOT_DIR}/perf_test/sql/cleanup_test_data.sql"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "compose file not found: ${COMPOSE_FILE}" >&2
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: ${ENV_FILE}" >&2
  exit 1
fi

if [[ ! -f "$GENERATE_SQL" || ! -f "$CLEANUP_SQL" ]]; then
  echo "perf SQL files are missing" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

compose() {
  docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" "$@"
}

echo "[perf-test] cleaning previous perf dataset"
compose exec -T \
  -e PGPASSWORD="$POSTGRES_ADMIN_PASSWORD" \
  db psql \
  -v ON_ERROR_STOP=1 \
  -U "$POSTGRES_ADMIN_USER" \
  -d "$POSTGRES_DB" \
  -f - < "$CLEANUP_SQL"

echo "[perf-test] generating dataset: movies=${MOVIE_COUNT}, users=${USER_COUNT}, actors=${ACTOR_COUNT}"
compose exec -T \
  -e PGPASSWORD="$POSTGRES_ADMIN_PASSWORD" \
  db psql \
  -v ON_ERROR_STOP=1 \
  -v movie_count="${MOVIE_COUNT}" \
  -v user_count="${USER_COUNT}" \
  -v actor_count="${ACTOR_COUNT}" \
  -v selection_count="${SELECTION_COUNT}" \
  -v movies_per_selection="${MOVIES_PER_SELECTION}" \
  -v interactions_per_movie="${INTERACTIONS_PER_MOVIE}" \
  -v watch_rows_per_user="${WATCH_ROWS_PER_USER}" \
  -U "$POSTGRES_ADMIN_USER" \
  -d "$POSTGRES_DB" \
  -f - < "$GENERATE_SQL"

echo "[perf-test] dataset is ready"
