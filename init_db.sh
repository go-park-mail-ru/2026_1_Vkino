#!/bin/bash

docker exec -i prod-db-1 psql -U vkino_user -d vkino < migrations/20260223210612_migration.up.sql

docker cp build/scripts/fill.sql prod-db-1:/tmp/fill.sql
docker exec prod-db-1 psql -U vkino_user -d vkino -f /tmp/fill.sql
