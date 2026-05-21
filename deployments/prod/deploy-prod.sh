#!/usr/bin/env bash

set -Eeuo pipefail

log() {
  printf '[deploy-prod] %s\n' "$*"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    log "required command not found: $1"
    exit 1
  }
}

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/deployments/prod/compose.yaml"
ENV_FILE="${ROOT_DIR}/deployments/prod/.env"
LOCK_FILE="${ROOT_DIR}/deployments/prod/.deploy-prod.lock"
PROJECT_NAME="${COMPOSE_PROJECT_NAME:-prod}"
ROLLOUT_TIMEOUT_SECONDS=120
ROLLOUT_WAIT_AFTER_HEALTHY_SECONDS=5

require_cmd git
require_cmd docker
require_cmd flock

if [[ ! -f "${COMPOSE_FILE}" ]]; then
  log "compose file not found: ${COMPOSE_FILE}"
  exit 1
fi

if [[ ! -f "${ENV_FILE}" ]]; then
  log "env file not found: ${ENV_FILE}"
  exit 1
fi

if ! docker rollout --help >/dev/null 2>&1; then
  log "docker rollout is not installed. See docs/deployment/zero-downtime-rollout.md"
  exit 1
fi

exec 9>"${LOCK_FILE}"
if ! flock -n 9; then
  log "another production deployment is already running"
  exit 1
fi

cd "${ROOT_DIR}"

compose() {
  docker compose \
    --project-name "${PROJECT_NAME}" \
    -f "${COMPOSE_FILE}" \
    --env-file "${ENV_FILE}" \
    "$@"
}

rollout() {
  local service="$1"

  docker rollout \
    --project-name "${PROJECT_NAME}" \
    -f "${COMPOSE_FILE}" \
    --env-file "${ENV_FILE}" \
    --timeout "${ROLLOUT_TIMEOUT_SECONDS}" \
    --wait-after-healthy "${ROLLOUT_WAIT_AFTER_HEALTHY_SECONDS}" \
    "${service}"
}

log "fetching latest main branch"
git fetch origin main
git reset --hard origin/main

export APP_VERSION="${APP_VERSION:-$(git rev-parse --short HEAD)}"
log "using app version ${APP_VERSION}"

log "building application images on the production server"
compose build auth-service user-service movie-service party-service api-gateway

log "running database migrations"
compose run --rm migrate

for service in \
  auth-service \
  user-service \
  movie-service \
  party-service \
  api-gateway
do
  log "rolling out ${service}"
  rollout "${service}"
done

log "current compose status"
compose ps

log "pruning dangling images"
docker image prune -f >/dev/null

log "deployment completed successfully"
