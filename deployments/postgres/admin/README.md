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

- `../init-db.sh`
  Оркестрирует полный init flow для `dev` и `prod`:
  1. поднимает `db` c `--force-recreate`
  2. ждёт `healthy`
  3. находит docker network контейнера `db`
  4. запускает `bootstrap.sh` во временном `postgres:18.3` контейнере
  5. запускает `migrate`
  6. если миграции падают с `permission denied`, автоматически применяет fallback grants и повторяет миграции
  7. запускает `apply-runtime-grants.sh`

## Ручной порядок запуска

Основной способ запуска из корня репозитория:

```bash
make init-db
```

Для `prod`:

```bash
make init-db DEPLOY_ENV=prod
```

Скрипт также можно вызвать напрямую:

```bash
./deployments/postgres/init-db.sh dev
```

```bash
./deployments/postgres/init-db.sh prod
```

## Логи

Основной ход инициализации печатается прямо в stdout `make init-db` или `init-db.sh`.
Для service-логов перейдите в нужную deployment-директорию. Для `prod` используйте `deployments/prod` вместо `deployments/dev`.

```bash
cd deployments/dev
docker compose logs db
```

## Проверка результата

Из нужной deployment-директории. Для `prod` используйте `deployments/prod` вместо `deployments/dev`.

```bash
cd deployments/dev
docker compose ps -a
```

```bash
cd deployments/dev
docker compose exec db psql -U "$POSTGRES_ADMIN_USER" -d "$POSTGRES_DB" -c "
SELECT rolname, rolsuper, rolcreatedb, rolcreaterole, rolreplication, rolconnlimit
FROM pg_roles
WHERE rolname LIKE 'vkino_%'
ORDER BY rolname;
"
```

Перед запуском нужно заполнить `.env` в `deployments/dev` или `deployments/prod` по соответствующему `.env.example`.
