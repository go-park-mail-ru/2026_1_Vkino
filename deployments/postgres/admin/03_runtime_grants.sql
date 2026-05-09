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

SELECT format('GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE user_interaction TO %I', :'user_user')
\gexec

SELECT format('GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE user_interaction_review_reaction TO %I', :'user_user')
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
    'GRANT SELECT ON TABLE users, movie, episode, actor, genre, language, country, user_interaction, user_interaction_review_reaction, movie_external_rating, genre_to_movie, actor_to_movie, selection, movie_to_selection TO %I',
    :'movie_user'
)
\gexec

SELECT format('GRANT SELECT ON TABLE users TO %I', :'party_user')
\gexec

SELECT format(
    'GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE vkino_room, vkino_room_member, vkino_room_invite, vkino_room_playback_state TO %I',
    :'party_user'
)
\gexec

SELECT format(
    'GRANT SELECT, INSERT ON TABLE vkino_room_chat_message, vkino_room_chat_bet, vkino_room_chat_bet_variant, vkino_room_chat_bet_answer TO %I',
    :'party_user'
)
\gexec

SELECT format('GRANT SELECT, INSERT, UPDATE ON TABLE watch_progress_episode TO %I', :'movie_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON SCHEMA public TO %I', :'migrator_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %I', :'migrator_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %I', :'migrator_user')
\gexec

SELECT format('GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO %I', :'migrator_user')
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

SELECT format('GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %I', :'party_user')
\gexec
