\set ON_ERROR_STOP on

-- VKino DB service roles
--
-- Этот скрипт выполняется администратором БД, НЕ runtime-приложением.
-- Он создаёт отдельные сервисные роли для микросервисов и выдаёт им
-- минимально необходимые права.
--
-- Запускать примерно так:
--
-- psql \
--   -h localhost \
--   -U vkino_admin \
--   -d vkino \
--   -v auth_password="'strong_auth_password'" \
--   -v user_password="'strong_user_password'" \
--   -v movie_password="'strong_movie_password'" \
--   -v migrator_password="'strong_migrator_password'" \
--   -v monitoring_password="'strong_monitoring_password'" \
--   -f deployments/postgres/admin/01_service_roles.sql


-- 1. Runtime-роли приложения

SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'vkino_auth_service')
        THEN format(
            'CREATE ROLE vkino_auth_service LOGIN PASSWORD %s NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 20',
            :'auth_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

ALTER ROLE vkino_auth_service
    WITH LOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION
    CONNECTION LIMIT 20
    PASSWORD :auth_password;


SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'vkino_user_service')
        THEN format(
            'CREATE ROLE vkino_user_service LOGIN PASSWORD %s NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 30',
            :'user_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

ALTER ROLE vkino_user_service
    WITH LOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION
    CONNECTION LIMIT 30
    PASSWORD :user_password;


SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'vkino_movie_service')
        THEN format(
            'CREATE ROLE vkino_movie_service LOGIN PASSWORD %s NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 40',
            :'movie_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

ALTER ROLE vkino_movie_service
    WITH LOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION
    CONNECTION LIMIT 40
    PASSWORD :movie_password;


-- 2. Роль для миграций

SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'vkino_migrator')
        THEN format(
            'CREATE ROLE vkino_migrator LOGIN PASSWORD %s NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 5',
            :'migrator_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

ALTER ROLE vkino_migrator
    WITH LOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION
    CONNECTION LIMIT 5
    PASSWORD :migrator_password;


-- 3. Роль для мониторинга PostgreSQL

SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'vkino_monitoring')
        THEN format(
            'CREATE ROLE vkino_monitoring LOGIN PASSWORD %s NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 5',
            :'monitoring_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

ALTER ROLE vkino_monitoring
    WITH LOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION
    CONNECTION LIMIT 5
    PASSWORD :monitoring_password;

GRANT pg_monitor TO vkino_monitoring;


-- 4. Закрытие лишних прав

REVOKE ALL ON DATABASE vkino FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

GRANT CONNECT ON DATABASE vkino TO
    vkino_auth_service,
    vkino_user_service,
    vkino_movie_service,
    vkino_migrator,
    vkino_monitoring;

GRANT USAGE ON SCHEMA public TO
    vkino_auth_service,
    vkino_user_service,
    vkino_movie_service,
    vkino_migrator,
    vkino_monitoring;


-- ----------------------------------------------------------------------------
-- 5. Права vkino_auth_service
--
-- auth-service:
--   - регистрация пользователя;
--   - вход по email;
--   - обновление пароля;
--   - создание, чтение и удаление refresh-сессии.
-- ----------------------------------------------------------------------------

GRANT SELECT, INSERT, UPDATE
ON TABLE users
TO vkino_auth_service;

GRANT SELECT, INSERT, UPDATE, DELETE
ON TABLE user_session
TO vkino_auth_service;


-- ----------------------------------------------------------------------------
-- 6. Права vkino_user_service
--
-- user-service:
--   - чтение/обновление профиля;
--   - избранное;
--   - друзья;
--   - заявки в друзья;
--   - support tickets.
-- ----------------------------------------------------------------------------

GRANT SELECT, UPDATE
ON TABLE users
TO vkino_user_service;

GRANT SELECT
ON TABLE movie
TO vkino_user_service;

GRANT SELECT, INSERT, UPDATE
ON TABLE user_interaction
TO vkino_user_service;

GRANT SELECT, INSERT, DELETE
ON TABLE friend
TO vkino_user_service;

GRANT SELECT, INSERT, UPDATE, DELETE
ON TABLE friend_request
TO vkino_user_service;

GRANT SELECT, INSERT, UPDATE
ON TABLE support_ticket
TO vkino_user_service;

GRANT SELECT, INSERT
ON TABLE support_ticket_message
TO vkino_user_service;


-- ----------------------------------------------------------------------------
-- 7. Права vkino_movie_service
--
-- movie-service:
--   - каталог фильмов;
--   - жанры;
--   - актёры;
--   - подборки;
--   - эпизоды;
--   - прогресс просмотра;
--   - проверка избранного.
-- ----------------------------------------------------------------------------

GRANT SELECT
ON TABLE
    movie,
    episode,
    actor,
    genre,
    language,
    country,
    genre_to_movie,
    actor_to_movie,
    selection,
    movie_to_selection
TO vkino_movie_service;

GRANT SELECT, INSERT, UPDATE
ON TABLE watch_progress_episode
TO vkino_movie_service;

GRANT SELECT
ON TABLE user_interaction
TO vkino_movie_service;


-- ----------------------------------------------------------------------------
-- 8. Права vkino_migrator
--
-- Мигратор должен уметь менять схему.
-- В production это всё равно не runtime-пользователь приложения.
-- ----------------------------------------------------------------------------

GRANT ALL PRIVILEGES ON SCHEMA public TO vkino_migrator;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO vkino_migrator;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO vkino_migrator;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO vkino_migrator;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL PRIVILEGES ON TABLES TO vkino_migrator;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL PRIVILEGES ON SEQUENCES TO vkino_migrator;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL PRIVILEGES ON FUNCTIONS TO vkino_migrator;


-- ----------------------------------------------------------------------------
-- 9. Доступ к sequence для identity-колонок
--
-- Runtime-сервисам нужен USAGE на sequence тех таблиц, куда они делают INSERT.
-- Здесь можно дать на все sequence в public: это не даёт прав писать в таблицы,
-- но позволяет корректно работать INSERT в разрешённые таблицы.
-- ----------------------------------------------------------------------------

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO
    vkino_auth_service,
    vkino_user_service,
    vkino_movie_service;