\set ON_ERROR_STOP on

-- Pre-migration PostgreSQL administrative setup.
-- This script creates service roles, base connectivity grants and migrator schema access.

SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'auth_user')
        THEN format(
            'CREATE ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 20',
            :'auth_user',
            :'auth_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

SELECT format(
    'ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 20 PASSWORD %L',
    :'auth_user',
    :'auth_password'
)
\gexec

SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'user_user')
        THEN format(
            'CREATE ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 30',
            :'user_user',
            :'user_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

SELECT format(
    'ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 30 PASSWORD %L',
    :'user_user',
    :'user_password'
)
\gexec

SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'movie_user')
        THEN format(
            'CREATE ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 40',
            :'movie_user',
            :'movie_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

SELECT format(
    'ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 40 PASSWORD %L',
    :'movie_user',
    :'movie_password'
)
\gexec

SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'migrator_user')
        THEN format(
            'CREATE ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 5',
            :'migrator_user',
            :'migrator_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

SELECT format(
    'ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 5 PASSWORD %L',
    :'migrator_user',
    :'migrator_password'
)
\gexec

SELECT
    CASE
        WHEN NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'monitoring_user')
        THEN format(
            'CREATE ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 5',
            :'monitoring_user',
            :'monitoring_password'
        )
        ELSE 'SELECT 1'
    END
\gexec

SELECT format(
    'ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION CONNECTION LIMIT 5 PASSWORD %L',
    :'monitoring_user',
    :'monitoring_password'
)
\gexec

SELECT format('GRANT pg_monitor TO %I', :'monitoring_user')
\gexec

SELECT format('REVOKE ALL ON DATABASE %I FROM PUBLIC', :'db_name')
\gexec

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

SELECT format('GRANT CONNECT ON DATABASE %I TO %I', :'db_name', :'auth_user')
\gexec

SELECT format('GRANT CONNECT ON DATABASE %I TO %I', :'db_name', :'user_user')
\gexec

SELECT format('GRANT CONNECT ON DATABASE %I TO %I', :'db_name', :'movie_user')
\gexec

SELECT format('GRANT CONNECT ON DATABASE %I TO %I', :'db_name', :'migrator_user')
\gexec

SELECT format('GRANT CONNECT ON DATABASE %I TO %I', :'db_name', :'monitoring_user')
\gexec

SELECT format('GRANT USAGE ON SCHEMA public TO %I', :'auth_user')
\gexec

SELECT format('GRANT USAGE ON SCHEMA public TO %I', :'user_user')
\gexec

SELECT format('GRANT USAGE ON SCHEMA public TO %I', :'movie_user')
\gexec

SELECT format('GRANT USAGE ON SCHEMA public TO %I', :'migrator_user')
\gexec

SELECT format('GRANT USAGE ON SCHEMA public TO %I', :'monitoring_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON SCHEMA public TO %I', :'migrator_user')
\gexec
