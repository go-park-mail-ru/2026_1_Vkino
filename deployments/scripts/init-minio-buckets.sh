#!/bin/sh
set -e

alias_name="${MINIO_ALIAS:-local}"
minio_url="${MINIO_URL:-http://minio:9000}"

until /usr/bin/mc alias set "$alias_name" "$minio_url" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"; do
	echo "waiting for minio..."
	sleep 2
done

buckets="
vkino-actors
vkino-posters
vkino-cards
vkino-avatars
vkino-support
vkino-videos
"

for bucket in $buckets; do
	/usr/bin/mc mb -p "$alias_name/$bucket" || true
	/usr/bin/mc anonymous set none "$alias_name/$bucket" || true
done
