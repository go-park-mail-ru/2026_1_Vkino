# PostgreSQL Admin Scripts

Эта директория содержит административные скрипты PostgreSQL для VKino.
Они отделены от обычных миграций приложения, потому что роли, базовые права,
расширения и runtime-grants применяются на разных этапах жизненного цикла БД.

## Зачем этапы разделены

Используется три последовательных этапа:

1. `01_admin_setup.sql`
   Создаёт сервисные роли, базовые права подключения и права migrator на схему.
2. Обычные миграции приложения
   Создают таблицы, sequence, индексы и остальную схему.
3. `03_runtime_grants.sql`
   Выдаёт runtime-права на уже существующие таблицы и sequence.

Такое разделение нужно потому, что права на таблицы нельзя корректно выдавать до
миграций: на pre-migration этапе нужных таблиц может ещё не существовать.

## Что делает каждый файл

- `01_admin_setup.sql`
  Pre-migration действия: роли, `GRANT CONNECT`, `GRANT USAGE`, базовые revoke.
- `02_extensions.sql`
  Создаёт административные расширения, нужные инстансу PostgreSQL.
- `03_runtime_grants.sql`
  Post-migration права на таблицы, sequence и объекты схемы.

## Shell-скрипты

- `bootstrap.sh`
  Подключается к `db` под `POSTGRES_ADMIN_USER`, выполняет:
  1. `02_extensions.sql`
  2. `01_admin_setup.sql`

- `apply-runtime-grants.sh`
  Подключается к `db` под `POSTGRES_ADMIN_USER` и выполняет:
  1. `03_runtime_grants.sql`

## Ручной порядок запуска

Запускать из директории нужного deployment (`deployments/dev` или `deployments/prod`):

```bash
docker compose up -d db
docker compose up postgres-bootstrap
docker compose up migrate
docker compose up postgres-grants
docker compose up -d
```

## Логи

```bash
docker compose logs postgres-bootstrap
docker compose logs migrate
docker compose logs postgres-grants
```

## Проверка результата

```bash
docker compose ps -a
```

```bash
docker compose exec db psql -U "$POSTGRES_ADMIN_USER" -d "$POSTGRES_DB" -c "
SELECT rolname, rolsuper, rolcreatedb, rolcreaterole, rolreplication, rolconnlimit
FROM pg_roles
WHERE rolname LIKE 'vkino_%'
ORDER BY rolname;
"
```

## Почему legacy script больше не используется

Старый combined roles/grants script перенесён в `legacy/`.
Он смешивал pre-migration действия и post-migration права на таблицы, из-за чего
deployment-цепочка становилась неочевидной и более хрупкой.
