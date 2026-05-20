#!/usr/bin/env bash

set -Eeuo pipefail

DEPLOY_ENV="${1:-dev}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
LOAD_USER_EMAIL="${LOAD_USER_EMAIL:-perf_load_user@example.test}"
LOAD_USER_PASSWORD="${LOAD_USER_PASSWORD:-PerfLoadPass123!}"

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/deployments/${DEPLOY_ENV}/compose.yaml"
ENV_FILE="${ROOT_DIR}/deployments/${DEPLOY_ENV}/.env"
SETUP_SQL="${ROOT_DIR}/perf_test/sql/setup_load_user_state.sql"

if [[ ! -f "$COMPOSE_FILE" || ! -f "$ENV_FILE" || ! -f "$SETUP_SQL" ]]; then
  echo "perf load-user setup prerequisites are missing" >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required to parse auth JSON responses" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

compose() {
  docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" "$@"
}

signin_body="$(mktemp)"
signup_body="$(mktemp)"
trap 'rm -f "$signin_body" "$signup_body"' EXIT

signin_code="$(
  curl -sS -o "$signin_body" -w '%{http_code}' \
    -X POST "${BASE_URL}/user/sign-in" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${LOAD_USER_EMAIL}\",\"password\":\"${LOAD_USER_PASSWORD}\"}"
)"

access_token=""
if [[ "$signin_code" == "200" ]]; then
  access_token="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["access_token"])' "$signin_body")"
else
  signup_code="$(
    curl -sS -o "$signup_body" -w '%{http_code}' \
      -X POST "${BASE_URL}/user/sign-up" \
      -H "Content-Type: application/json" \
      -d "{\"email\":\"${LOAD_USER_EMAIL}\",\"password\":\"${LOAD_USER_PASSWORD}\"}"
  )"

  if [[ "$signup_code" != "201" ]]; then
    echo "failed to sign in or sign up load user" >&2
    echo "sign-in status: ${signin_code}" >&2
    echo "sign-up status: ${signup_code}" >&2
    exit 1
  fi

  access_token="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["access_token"])' "$signup_body")"
fi

compose exec -T \
  -e PGPASSWORD="$POSTGRES_ADMIN_PASSWORD" \
  db psql \
  -v ON_ERROR_STOP=1 \
  -v "load_user_email=${LOAD_USER_EMAIL}" \
  -U "$POSTGRES_ADMIN_USER" \
  -d "$POSTGRES_DB" \
  -f - < "$SETUP_SQL" >/dev/null

setup_counts="$(
  compose exec -T \
    -e PGPASSWORD="$POSTGRES_ADMIN_PASSWORD" \
    db psql \
    -Atq \
    -U "$POSTGRES_ADMIN_USER" \
    -d "$POSTGRES_DB" \
    -c "
      select
        count(*) filter (where ui.is_favorite = true),
        count(distinct wpe.episode_id)
      from users u
      left join user_interaction ui on ui.user_id = u.id
      left join watch_progress_episode wpe on wpe.user_id = u.id
      where u.email = '${LOAD_USER_EMAIL}';
    "
)"

IFS='|' read -r favorite_count watch_count <<< "$setup_counts"

if [[ "${favorite_count:-0}" == "0" || "${watch_count:-0}" == "0" ]]; then
  echo "load user state was not prepared correctly: favorites=${favorite_count:-0}, watch_rows=${watch_count:-0}" >&2
  exit 1
fi

printf '%s\n' "$access_token"
