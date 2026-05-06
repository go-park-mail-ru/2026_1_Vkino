\set ON_ERROR_STOP on

-- Post-migration runtime grants for application tables and sequences.

SELECT format('GRANT SELECT, INSERT, UPDATE ON TABLE users TO %I', :'auth_user')
\gexec

SELECT format('GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE user_session TO %I', :'auth_user')
\gexec

SELECT format('GRANT SELECT, UPDATE ON TABLE users TO %I', :'user_user')
\gexec

SELECT format('GRANT SELECT ON TABLE movie TO %I', :'user_user')
\gexec

SELECT format('GRANT SELECT, INSERT, UPDATE ON TABLE user_interaction TO %I', :'user_user')
\gexec

SELECT format('GRANT SELECT, INSERT, DELETE ON TABLE friend TO %I', :'user_user')
\gexec

SELECT format('GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE friend_request TO %I', :'user_user')
\gexec

SELECT format('GRANT SELECT, INSERT, UPDATE ON TABLE support_ticket TO %I', :'user_user')
\gexec

SELECT format('GRANT SELECT, INSERT ON TABLE support_ticket_message TO %I', :'user_user')
\gexec

SELECT format(
    'GRANT SELECT ON TABLE movie, episode, actor, genre, language, country, genre_to_movie, actor_to_movie, selection, movie_to_selection TO %I',
    :'movie_user'
)
\gexec

SELECT format('GRANT SELECT, INSERT, UPDATE ON TABLE watch_progress_episode TO %I', :'movie_user')
\gexec

SELECT format('GRANT SELECT ON TABLE user_interaction TO %I', :'movie_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON SCHEMA public TO %I', :'migrator_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %I', :'migrator_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %I', :'migrator_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO %I', :'migrator_user')
\gexec

SELECT format('GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %I', :'auth_user')
\gexec

SELECT format('GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %I', :'user_user')
\gexec

SELECT format('GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %I', :'movie_user')
\gexec
