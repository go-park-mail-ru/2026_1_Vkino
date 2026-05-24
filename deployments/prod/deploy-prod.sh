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
DEPLOY_BRANCH="${DEPLOY_BRANCH:-master}"
ROLLOUT_TIMEOUT_SECONDS="${ROLLOUT_TIMEOUT_SECONDS:-120}"
ROLLOUT_WAIT_AFTER_HEALTHY_SECONDS="${ROLLOUT_WAIT_AFTER_HEALTHY_SECONDS:-5}"

APP_SERVICES=(
  auth-service
  user-service
  movie-service
  party-service
  payment-service
  api-gateway
)

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

add_service() {
  local service="$1"
  local existing

  for existing in "${SERVICES_TO_DEPLOY[@]:-}"; do
    if [[ "${existing}" == "${service}" ]]; then
      return 0
    fi
  done

  SERVICES_TO_DEPLOY+=("${service}")
}

add_all_services() {
  local service

  for service in "${APP_SERVICES[@]}"; do
    add_service "${service}"
  done
}

path_affects_service() {
  local path="$1"

  case "${path}" in
    go.mod|go.sum|Makefile)
      add_all_services
      ;;

    pkg/*|build/*|proto/*)
      add_all_services
      ;;

    cmd/auth/*|cmd/auth/**|internal/app/auth-service/*|internal/app/auth-service/**|deployments/prod/auth/*|deployments/prod/auth/**)
      add_service auth-service
      ;;

    cmd/user/*|cmd/user/**|internal/app/user-service/*|internal/app/user-service/**|deployments/prod/user/*|deployments/prod/user/**)
      add_service user-service
      ;;

    cmd/movie/*|cmd/movie/**|internal/app/movie-service/*|internal/app/movie-service/**|deployments/prod/movie/*|deployments/prod/movie/**)
      add_service movie-service
      ;;

    cmd/party/*|cmd/party/**|internal/app/party-service/*|internal/app/party-service/**|deployments/prod/party/*|deployments/prod/party/**)
      add_service party-service
      ;;

    cmd/payment/*|cmd/payment/**|internal/app/payment-service/*|internal/app/payment-service/**|deployments/prod/payment/*|deployments/prod/payment/**)
      add_service payment-service
      ;;

    cmd/gateway/*|cmd/gateway/**|internal/app/api-gateway/*|internal/app/api-gateway/**|deployments/prod/gateway/*|deployments/prod/gateway/**)
      add_service api-gateway
      ;;

    deployments/prod/compose.yaml|deployments/prod/deploy-prod.sh)
      add_all_services
      ;;

    migrations/*|migrations/**)
      MIGRATIONS_CHANGED=1
      ;;

    deployments/prod/nginx/*|deployments/prod/nginx/**)
      NGINX_CHANGED=1
      ;;

    .github/workflows/*|.github/workflows/**|docs/*|docs/**|README.md)
      :
      ;;

    *)
      :
      ;;
  esac
}

log "fetching latest ${DEPLOY_BRANCH} branch"

OLD_REV="$(git rev-parse HEAD 2>/dev/null || true)"

git fetch origin "${DEPLOY_BRANCH}"
git reset --hard "origin/${DEPLOY_BRANCH}"

NEW_REV="$(git rev-parse HEAD)"

export APP_VERSION="${APP_VERSION:-$(git rev-parse --short HEAD)}"
log "using app version ${APP_VERSION}"

SERVICES_TO_DEPLOY=()
MIGRATIONS_CHANGED=0
NGINX_CHANGED=0

if [[ -z "${OLD_REV}" ]]; then
  log "old revision is unknown, scheduling all services"
  add_all_services
else
  log "detecting changed files between ${OLD_REV} and ${NEW_REV}"

  while IFS= read -r changed_path; do
    [[ -z "${changed_path}" ]] && continue
    log "changed: ${changed_path}"
    path_affects_service "${changed_path}"
  done < <(git diff --name-only "${OLD_REV}" "${NEW_REV}")
fi

if [[ "${FORCE_ROLLOUT_ALL:-0}" == "1" ]]; then
  log "FORCE_ROLLOUT_ALL=1, scheduling all services"
  add_all_services
fi

if [[ "${#SERVICES_TO_DEPLOY[@]}" -eq 0 && "${MIGRATIONS_CHANGED}" -eq 0 && "${NGINX_CHANGED}" -eq 0 ]]; then
  log "no deployable changes detected"
  compose ps
  exit 0
fi

if [[ "${#SERVICES_TO_DEPLOY[@]}" -gt 0 ]]; then
  log "building changed application images: ${SERVICES_TO_DEPLOY[*]}"
  compose build "${SERVICES_TO_DEPLOY[@]}"
else
  log "no application image changes detected"
fi

if [[ "${MIGRATIONS_CHANGED}" == "1" || "${#SERVICES_TO_DEPLOY[@]}" -gt 0 ]]; then
  log "running database migrations"
  compose run --rm migrate
else
  log "migrations were not changed and no app services will be rolled out; skipping migrations"
fi

for service in "${SERVICES_TO_DEPLOY[@]}"; do
  log "rolling out ${service}"
  rollout "${service}"
done

if [[ "${NGINX_CHANGED}" == "1" ]]; then
  log "nginx config changed, reloading nginx"
  compose exec -T nginx nginx -s reload
fi

log "current compose status"
compose ps

log "pruning dangling images"
docker image prune -f >/dev/null

log "deployment completed successfully"