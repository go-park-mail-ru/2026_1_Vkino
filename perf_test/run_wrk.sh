#!/usr/bin/env bash

set -Eeuo pipefail

SCENARIO="${1:-selection-all}"
DEPLOY_ENV="${DEPLOY_ENV:-dev}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
THREADS="${THREADS:-4}"
CONNECTIONS="${CONNECTIONS:-32}"
DURATION="${DURATION:-20s}"

if ! command -v wrk >/dev/null 2>&1; then
  echo "wrk is not installed. Install it locally and rerun, for example: brew install wrk" >&2
  exit 1
fi

URL=""
AUTH_HEADER=""

case "$SCENARIO" in
  selection-all)
    URL="${BASE_URL}/movie/selection/all"
    ;;
  selection-one)
    URL="${BASE_URL}/movie/selection/PerfTest%20Selection%20010"
    ;;
  movie-by-id)
    URL="${BASE_URL}/movie/106734"
    ;;
  search-movie)
    URL="${BASE_URL}/movie/search?query=050000"
    ;;
  search-movie-broad)
    URL="${BASE_URL}/movie/search?query=perftest%20movie"
    ;;
  favorites)
    URL="${BASE_URL}/user/favorites?limit=10&offset=0"
    AUTH_HEADER="Authorization: Bearer $(./perf_test/setup_load_user.sh "${DEPLOY_ENV}")"
    ;;
  watch-continue)
    URL="${BASE_URL}/user/watch/continue?limit=5"
    AUTH_HEADER="Authorization: Bearer $(./perf_test/setup_load_user.sh "${DEPLOY_ENV}")"
    ;;
  watch-history)
    URL="${BASE_URL}/user/watch/history?limit=10"
    AUTH_HEADER="Authorization: Bearer $(./perf_test/setup_load_user.sh "${DEPLOY_ENV}")"
    ;;
  *)
    echo "unknown scenario: ${SCENARIO}" >&2
    echo "available: selection-all, selection-one, movie-by-id, search-movie, search-movie-broad, favorites, watch-continue, watch-history" >&2
    exit 1
    ;;
esac

echo "[wrk] scenario=${SCENARIO} url=${URL} threads=${THREADS} connections=${CONNECTIONS} duration=${DURATION}"

if [[ -n "$AUTH_HEADER" ]]; then
  wrk -t"${THREADS}" -c"${CONNECTIONS}" -d"${DURATION}" -H "$AUTH_HEADER" "$URL"
else
  wrk -t"${THREADS}" -c"${CONNECTIONS}" -d"${DURATION}" "$URL"
fi
