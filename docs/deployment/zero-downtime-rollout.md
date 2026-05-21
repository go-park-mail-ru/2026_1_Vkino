# Zero-Downtime Rollout

Эта схема нужна для production-выкатки `Vkino` без `docker compose down`. Билд выполняется прямо на production-сервере, а `docker rollout` обновляет только application-сервисы:

- `auth-service`
- `user-service`
- `movie-service`
- `party-service`
- `api-gateway`

Через rollout не выкатываются:

- `nginx`
- `db`
- `minio`
- `minio-init`
- `migrate`
- `prometheus`
- `grafana`
- `cadvisor`
- `node-exporter`
- `postgres-exporter`

## Что требуется на сервере

На сервере уже должны быть установлены:

- Docker Engine
- Docker Compose v2
- `git`
- `flock`

### Установка docker-rollout

По официальной документации `docker-rollout` ставится как Docker CLI plugin:

```bash
sudo mkdir -p /usr/local/lib/docker/cli-plugins
sudo curl -fsSL \
  https://raw.githubusercontent.com/wowu/docker-rollout/main/docker-rollout \
  -o /usr/local/lib/docker/cli-plugins/docker-rollout
sudo chmod +x /usr/local/lib/docker/cli-plugins/docker-rollout
docker rollout --help
```

Если rollout нужен только одному пользователю, вместо системного пути можно использовать `~/.docker/cli-plugins/docker-rollout`.

## Что нужно подготовить в GitHub

В репозитории должны быть заведены Secrets:

- `PROD_HOST`
- `PROD_USER`
- `PROD_SSH_KEY`
- `PROD_SSH_PORT`
- `PROD_PROJECT_PATH`

`PROD_PROJECT_PATH` должен указывать на директорию проекта на production-сервере, где лежит этот репозиторий и файл `deployments/prod/.env`.

## Первый запуск на сервере

Один раз на сервере нужно:

1. Клонировать репозиторий.
2. Создать `deployments/prod/.env` и заполнить production-переменные.
3. Убедиться, что сертификаты и production-конфиги nginx уже лежат на сервере.
4. Сделать deploy-скрипт исполняемым:

```bash
chmod +x deployments/prod/deploy-prod.sh
```

5. Поднять non-rollout инфраструктуру:

```bash
docker compose -f deployments/prod/compose.yaml --env-file deployments/prod/.env up -d \
  db minio minio-init nginx prometheus grafana cadvisor node-exporter postgres-exporter
```

После этого application-сервисы можно выкатывать через rollout.

## Как работает выкладка

Workflow `.github/workflows/prod-deploy.yml` запускается на `push` в `main` и по SSH выполняет:

```bash
cd "$PROD_PROJECT_PATH"
./deployments/prod/deploy-prod.sh
```

Сам скрипт:

1. Берёт lock через `flock`, чтобы деплои не пересекались.
2. Делает `git fetch origin main`.
3. Делает `git reset --hard origin/main`.
4. Вычисляет `APP_VERSION` из короткого SHA коммита.
5. Собирает app images прямо на production-сервере.
6. Запускает `migrate` перед rollout.
7. По очереди выкатывает:
   - `auth-service`
   - `user-service`
   - `movie-service`
   - `party-service`
   - `api-gateway`
8. Показывает `docker compose ps`.

Важно: миграции перед rollout должны быть backward-compatible. Новый код должен уметь работать со схемой в переходном состоянии, пока часть контейнеров уже обновилась, а часть ещё нет.

## Ручная проверка rollout

Проверить, что plugin установлен:

```bash
docker rollout --help
```

Собрать свежие образы и прогнать миграции вручную:

```bash
docker compose -f deployments/prod/compose.yaml --env-file deployments/prod/.env build \
  auth-service user-service movie-service party-service api-gateway

docker compose -f deployments/prod/compose.yaml --env-file deployments/prod/.env run --rm migrate
```

Проверить rollout по одному сервису:

```bash
docker rollout \
  -f deployments/prod/compose.yaml \
  --env-file deployments/prod/.env \
  --project-name prod \
  --timeout 120 \
  --wait-after-healthy 5 \
  api-gateway
```

Посмотреть состояние контейнеров:

```bash
docker compose -f deployments/prod/compose.yaml --env-file deployments/prod/.env ps
```

## Что важно про трафик

- `nginx` остаётся отдельным reverse proxy и не выкатывается через rollout.
- `/api/` теперь проксируется через Docker DNS re-resolve, чтобы nginx подхватывал новый контейнер `api-gateway` во время rollout.
- `db` и `minio` не входят в rollout-цепочку.
- WebSocket-соединения могут переподключиться во время смены контейнера `api-gateway`, но HTTP API должен оставаться доступным.

## Источники

- docker-rollout docs: https://docker-rollout.wowu.dev/
- docker-rollout CLI options: https://docker-rollout.wowu.dev/cli-options.html
- docker-rollout GitHub: https://github.com/wowu/docker-rollout
- appleboy ssh-action releases: https://github.com/appleboy/ssh-action/releases
