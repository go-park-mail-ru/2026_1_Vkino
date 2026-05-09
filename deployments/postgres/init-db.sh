#!/usr/bin/env bash

set -Eeuo pipefail

POSTGRES_IMAGE="postgres:18.3"
DEPLOY_ENV="${1:-dev}"

case "$DEPLOY_ENV" in
  dev|prod)
    ;;
  *)
    echo "usage: $0 [dev|prod]" >&2
    exit 1
    ;;
esac

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_DIR="${ROOT_DIR}/deployments/${DEPLOY_ENV}"
COMPOSE_FILE="${COMPOSE_DIR}/compose.yaml"
ENV_FILE="${COMPOSE_DIR}/.env"
ADMIN_DIR="${ROOT_DIR}/deployments/postgres/admin"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "compose file not found: ${COMPOSE_FILE}" >&2
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: ${ENV_FILE}" >&2
  echo "create it from ${COMPOSE_DIR}/.env.example first" >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is not installed or not found in PATH" >&2
  exit 1
fi

if docker info >/dev/null 2>&1; then
  DOCKER_CMD=(docker)
elif command -v sudo >/dev/null 2>&1; then
  DOCKER_CMD=(sudo docker)
else
  echo "docker requires elevated privileges and sudo is unavailable" >&2
  exit 1
fi

MIGRATE_LOG=""

cleanup() {
  if [[ -n "$MIGRATE_LOG" && -f "$MIGRATE_LOG" ]]; then
    rm -f "$MIGRATE_LOG"
  fi
}

trap cleanup EXIT

log() {
  printf '[init-db][%s] %s\n' "$DEPLOY_ENV" "$1"
}

compose() {
  "${DOCKER_CMD[@]}" compose \
    -f "$COMPOSE_FILE" \
    --env-file "$ENV_FILE" \
    "$@"
}

load_env() {
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
}

wait_for_db() {
  local container_id=""
  local status=""
  local attempts=0

  while (( attempts < 60 )); do
    container_id="$(compose ps -q db)"
    if [[ -n "$container_id" ]]; then
      status="$("${DOCKER_CMD[@]}" inspect "$container_id" --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' 2>/dev/null || true)"
      if [[ "$status" == "healthy" ]]; then
        return 0
      fi
      if [[ "$status" == "unhealthy" ]]; then
        log "db became unhealthy, printing logs"
        compose logs db || true
        return 1
      fi
    fi
    sleep 2
    attempts=$((attempts + 1))
  done

  log "timed out waiting for db to become healthy"
  compose ps
  return 1
}

resolve_network() {
  local container_id
  local network_name

  container_id="$(compose ps -q db)"
  if [[ -z "$container_id" ]]; then
    echo "db container is not running" >&2
    return 1
  fi

  network_name="$("${DOCKER_CMD[@]}" inspect "$container_id" --format '{{range $name, $_ := .NetworkSettings.Networks}}{{println $name}}{{end}}' | awk 'NF {print; exit}')"
  if [[ -z "$network_name" ]]; then
    echo "failed to resolve docker network for db" >&2
    return 1
  fi

  printf '%s\n' "$network_name"
}

run_bootstrap() {
  local network_name="$1"

  log "running postgres bootstrap scripts"
  "${DOCKER_CMD[@]}" run --rm \
    --network "$network_name" \
    --env-file "$ENV_FILE" \
    -v "${ADMIN_DIR}:/scripts:ro" \
    "$POSTGRES_IMAGE" \
    /bin/sh /scripts/bootstrap.sh
}

grant_migrator_permissions() {
  log "applying fallback grants for migrator user"
  load_env

  compose exec -T \
    -e PGPASSWORD="$POSTGRES_ADMIN_PASSWORD" \
    db psql \
    -v ON_ERROR_STOP=1 \
    -U "$POSTGRES_ADMIN_USER" \
    -d "$POSTGRES_DB" \
    -c "
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ${VKINO_MIGRATOR_DB_USER};
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${VKINO_MIGRATOR_DB_USER};
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO ${VKINO_MIGRATOR_DB_USER};
"
}

run_migrations() {
  MIGRATE_LOG="$(mktemp)"
  log "running migrations"

  if compose run --rm migrate 2>&1 | tee "$MIGRATE_LOG"; then
    rm -f "$MIGRATE_LOG"
    MIGRATE_LOG=""
    return 0
  fi

  if grep -qi "permission denied" "$MIGRATE_LOG"; then
    log "migration hit permission denied, retrying after fallback grants"
    grant_migrator_permissions
    compose run --rm migrate
    rm -f "$MIGRATE_LOG"
    MIGRATE_LOG=""
    return 0
  fi

  log "migration failed for a reason other than permission denied"
  return 1
}

run_runtime_grants() {
  local network_name="$1"

  log "applying runtime grants"
  "${DOCKER_CMD[@]}" run --rm \
    --network "$network_name" \
    --env-file "$ENV_FILE" \
    -v "${ADMIN_DIR}:/scripts:ro" \
    "$POSTGRES_IMAGE" \
    /bin/sh /scripts/apply-runtime-grants.sh
}

log "starting db container"
compose up -d --force-recreate db

log "waiting for db healthcheck"
wait_for_db

NETWORK_NAME="$(resolve_network)"
log "using docker network ${NETWORK_NAME}"

run_bootstrap "$NETWORK_NAME"
run_migrations
run_runtime_grants "$NETWORK_NAME"

log "database initialization completed successfully"
