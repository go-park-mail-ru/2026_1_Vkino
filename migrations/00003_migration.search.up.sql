BEGIN;

create index if not exists movie_search_idx
on movie
using gin ((
	setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
	setweight(to_tsvector('simple', coalesce(description, '')), 'B') ||
	setweight(to_tsvector('simple', coalesce(director, '')), 'C')
));

create index if not exists actor_search_idx
on actor
using gin ((
	setweight(to_tsvector('simple', coalesce(full_name, '')), 'A') ||
	setweight(to_tsvector('simple', coalesce(biography, '')), 'B')
));

COMMIT;
