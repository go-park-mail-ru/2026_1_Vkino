# Perf Test Report

## 1. Основная сущность

Основная сущность для оптимизации: `movie`.

Причина простая: почти все тяжёлые read-сценарии сходятся в неё напрямую или через связи:

- `GET /movie/{id}` тянет базу фильма, жанры, актёров, эпизоды, внешние рейтинги и отзывы.
- `GET /movie/selection/all` и `GET /movie/selection/{selection}` строят подборки из фильмов и считают рейтинг подборки.
- `GET /movie/search` использует полнотекстовый поиск по фильмам.
- `GET /user/favorites`, `GET /user/watch/continue`, `GET /user/watch/history` в итоге тоже возвращают карточки `movie`.

## 2. Как поднималось окружение

- Рабочий стек: `deployments/dev`.
- Базовая инициализация БД: `make init-db DEPLOY_ENV=dev`.
- Снимок DDL до оптимизаций лежит в [`perf_test/init.sql`](init.sql).
- Для честного сравнения я не менял набор данных между замерами, а переключал только [`migrations/00008_perf_optimizations.down.sql`](../migrations/00008_perf_optimizations.down.sql) и [`migrations/00008_perf_optimizations.up.sql`](../migrations/00008_perf_optimizations.up.sql).

## 3. Как генерировались данные

Генератор: [`perf_test/generate_data.sh`](generate_data.sh) + [`perf_test/sql/generate_test_data.sql`](sql/generate_test_data.sql).

Команда:

```bash
./perf_test/generate_data.sh dev
```

Сгенерированный набор:

| Сущность | Объём |
| --- | ---: |
| `movie` | 100000 |
| `actor` | 20000 |
| `users` | 20000 |
| `selection` | 40 |
| `movie_to_selection` | 16075 |
| `user_interaction` | 1200000 |
| `user_interaction_review_reaction` | 600000 |
| `episode` | 120000 |
| `watch_progress_episode` | 240000 |

Данные воспроизводимы: тестовые строки помечены префиксами `PerfTest ...` и `perf_test_...`.

Очистка лежит в [`perf_test/sql/cleanup_test_data.sql`](sql/cleanup_test_data.sql).

## 4. Инструмент нагрузки

Выбран `wrk`.

Причины:

- он хорошо подходит именно для быстрых повторных HTTP read-тестов;
- команды короткие и их удобно класть в README;
- overhead минимален.

Обёртка для запуска: [`perf_test/run_wrk.sh`](run_wrk.sh).

Подготовка авторизованного пользователя: [`perf_test/setup_load_user.sh`](setup_load_user.sh) + [`perf_test/sql/setup_load_user_state.sql`](sql/setup_load_user_state.sql).

Важно: в моём окружении бинарника `wrk` не было, поэтому сами SQL-оптимизации я валидировал через `EXPLAIN ANALYZE`, а HTTP-часть проверил до рабочего состояния подготовкой токена и реальными ответами endpoint'ов. Скрипт `run_wrk.sh` готов к запуску после `brew install wrk`.

## 5. Какие сценарии тестировались

- `GET /movie/selection/all`
- `GET /movie/selection/PerfTest Selection 010`
- `GET /movie/106734`
- `GET /movie/search?query=050000`
- `GET /movie/search?query=perftest movie`
- `GET /user/favorites?limit=10&offset=0`
- `GET /user/watch/continue?limit=5`
- `GET /user/watch/history?limit=10`

## 6. Итерация 1: подборки с фильмами и рейтингами

Затронутые SQL: `sqlGetAllSelectionMovies` и `sqlGetSelectionMoviesByTitle` в `internal/app/movie-service/repository/postgres/query.go`.

Исходная идея запроса была такой:

```sql
with movie_user_ratings as (
  select ui.movie_id, avg(ui.rating)
  from user_interaction ui
  where ui.rating is not null
  group by ui.movie_id
),
selection_ratings as (
  select ms.selection_id, avg(mur.avg_rating)
  from movie_to_selection ms
  left join movie_user_ratings mur on mur.movie_id = ms.movie_id
  group by ms.selection_id
)
...
```

Проблема: рейтинг пересчитывался по всем оценённым фильмам в системе на каждый запрос, даже если нужен один `selection` на 400 фильмов.

`EXPLAIN ANALYZE` до:

- `GET /movie/selection/all`: `Execution Time: 2691.418 ms`
- `GET /movie/selection/{selection}`: `Execution Time: 716.356 ms`
- ключевой узел плана: `Parallel Seq Scan on user_interaction` + `Finalize GroupAggregate` по почти всем оценкам

Изменение:

- сначала выделяю `selected_movies`;
- затем считаю `selected_movie_ids`;
- рейтинг агрегирую только по фильмам, которые реально входят в текущую подборку;
- добавил индексы `movie_to_selection_selection_id_id_movie_id_idx` и `user_interaction_movie_rating_idx`.

Новый каркас запроса:

```sql
with selected_movies as (...),
selected_movie_ids as (
  select distinct sm.movie_id
  from selected_movies sm
),
movie_user_ratings as (
  select ui.movie_id, avg(ui.rating)
  from selected_movie_ids sm
  join user_interaction ui on ui.movie_id = sm.movie_id
  where ui.rating is not null
  group by ui.movie_id
)
...
```

`EXPLAIN ANALYZE` после:

- `GET /movie/selection/all`: `Execution Time: 643.873 ms`
- `GET /movie/selection/{selection}`: `Execution Time: 36.032 ms`
- ключевой узел плана: `Index Only Scan using user_interaction_movie_rating_idx`

Вывод: materialized view здесь не понадобился. Подборки читаются сильно быстрее только за счёт ограничения множества фильмов и точечных индексов.

## 7. Итерация 2: отзывы в `/movie/{id}`

Затронутый SQL: `sqlGetMovieReviewsByMovieID` в `internal/app/movie-service/repository/postgres/query.go`.

Исходный запрос сначала агрегировал все реакции на все отзывы:

```sql
with reaction_counts as (
  select review_id, count(*) filter (...), count(*) filter (...)
  from user_interaction_review_reaction
  group by review_id
)
...
where ui.movie_id = $1
```

Проблема: для фильма с 12 отзывами запрос всё равно проходил по большому куску `user_interaction_review_reaction`.

`EXPLAIN ANALYZE` до:

- `Execution Time: 1049.587 ms`
- ключевой узел плана: `GroupAggregate` поверх `Index Scan using user_interaction_review_reaction_review_idx`, обработано около `500001` строк реакций

Изменение:

- сначала формирую `target_reviews` только для одного фильма;
- `reaction_counts` и `viewer_reactions` джойню только к `target_reviews`;
- добавил частичный covering-индекс `user_interaction_movie_review_idx`;
- predicate индекса повторяет исходную логику `blank comment != review`, чтобы не менять поведение API.

`EXPLAIN ANALYZE` после:

- `Execution Time: 9.883 ms`
- ключевой узел плана: `Index Only Scan using user_interaction_movie_review_idx`

Вывод: основная проблема была не в join'ах как таковых, а в неверном порядке вычислений. Сначала надо сузить множество отзывов, а уже потом трогать реакции.

## 8. Итерация 3: избранное и лишний round-trip

Затронутый SQL: `sqlGetFavorites` в `internal/app/user-service/repository/postgres/query.go` и метод `GetFavorites` в `internal/app/user-service/repository/postgres/user.go`.

До оптимизации было два похода в БД:

```sql
select ui.movie_id
from user_interaction ui
where ui.user_id = $1 and ui.is_favorite = true
order by ui.updated_at desc
limit $2 offset $3;

select count(*)
from user_interaction
where user_id = $1 and is_favorite = true;
```

Проблема:

- два отдельных round-trip;
- оба запроса сканировали `user_interaction` без подходящего индекса.

`EXPLAIN ANALYZE` до:

- page query: `Execution Time: 188.371 ms`
- count query: `Execution Time: 188.722 ms`
- ключевой узел плана: `Parallel Seq Scan on user_interaction`

Изменение:

- переписал на один запрос с `lateral`;
- тот же запрос сразу возвращает `movie_id` и `total_count`;
- добавил индекс `user_interaction_favorite_user_updated_idx`.

Новый запрос:

```sql
with total as (
  select count(*)::int as total_count
  from user_interaction ui
  where ui.user_id = $1 and ui.is_favorite = true
)
select p.movie_id, t.total_count
from total t
left join lateral (
  select ui.movie_id, ui.updated_at
  from user_interaction ui
  where ui.user_id = $1 and ui.is_favorite = true
  order by ui.updated_at desc
  limit $2 offset $3
) p on true
order by p.updated_at desc nulls last;
```

`EXPLAIN ANALYZE` после:

- единый запрос: `Execution Time: 3.158 ms`
- ключевой узел плана: два `Index Only Scan using user_interaction_favorite_user_updated_idx`

Вывод: здесь выигрыш дал и индекс, и удаление лишнего round-trip.

## 9. Итерация 4: `/movie/{id}` и связанные lookup'и

Проверил не только отзывы, но и lookup'и, которые строят карточку фильма.

Самый проблемный оказался запрос актёров:

```sql
select a.id, a.full_name, a.picture_file_key
from actor_to_movie am
join actor a on a.id = am.actor_id
where am.movie_id = $1
order by a.full_name;
```

`EXPLAIN ANALYZE` до:

- `Execution Time: 237.765 ms`
- ключевой узел плана: `Parallel Seq Scan on actor_to_movie`

После индекса `actor_to_movie_movie_id_actor_id_idx`:

- `Execution Time: 1.062 ms`
- ключевой узел плана: `Index Only Scan using actor_to_movie_movie_id_actor_id_idx`

Дополнительно индекс `genre_to_movie_movie_id_genre_id_idx` ускорил lookup жанров по `movie_id`, хотя там проблема была слабее.

## 10. Что проверил, но не стал менять

### Полнотекстовый поиск фильмов и актёров

Существующие индексы из [`migrations/00003_migration.search.up.sql`](../migrations/00003_migration.search.up.sql) уже полезны для селективных запросов.

Фактические результаты:

- movie search `query=050000`: `Execution Time: 3.334 ms`, используется `Bitmap Index Scan on movie_search_idx`
- actor search `query=005000`: `Execution Time: 0.628 ms`, используется `Bitmap Index Scan on actor_search_idx`
- broad movie search `query=perftest movie`: `Execution Time: 982.506 ms`, планировщик честно выбрал `Parallel Seq Scan`, потому что запрос матчит почти все `100000` тестовых фильмов
- broad actor search `query=perftest actor`: `Execution Time: 340.399 ms`, индекс используется, но возвращает сразу `20000` строк

Вывод: проблема broad search здесь в низкой селективности, а не в отсутствии индекса. Добавление ещё одного индекса или materialized view не убирает тот факт, что запросу всё равно нужно обработать почти весь набор.

### Continue watching / history

Проверил запросы `sqlGetContinueWatching` и `sqlGetWatchHistory` из `internal/app/movie-service/repository/postgres/query.go`.

На подготовленном пользователе:

- continue watching: `Execution Time: 7.071 ms`
- history: `Execution Time: 7.037 ms`

Вывод: bottleneck'ом они не были, поэтому запросы оставлены без переписывания.

## 11. Итоговые изменения

- Переписаны `sqlGetAllSelectionMovies`, `sqlGetSelectionMoviesByTitle`, `sqlGetMovieReviewsByMovieID`, `sqlGetFavorites`.
- Убран лишний count round-trip в `GetFavorites`.
- Добавлена миграция [`migrations/00008_perf_optimizations.up.sql`](../migrations/00008_perf_optimizations.up.sql) и rollback [`migrations/00008_perf_optimizations.down.sql`](../migrations/00008_perf_optimizations.down.sql).
- Добавлены индексы:
  - `actor_to_movie_movie_id_actor_id_idx`
  - `genre_to_movie_movie_id_genre_id_idx`
  - `movie_to_selection_selection_id_id_movie_id_idx`
  - `user_interaction_movie_rating_idx`
  - `user_interaction_movie_review_idx`
  - `user_interaction_favorite_user_updated_idx`
- Добавлен генератор `100k movie` и связанных сущностей.
- Добавлен безопасный cleanup тестовых данных.
- Добавлены `wrk`-сценарии и подготовка авторизованного load-user.

Views и materialized views не добавлял: после переписывания SQL и точечных индексов они уже не были нужны, а для рейтингов подборок ещё и ухудшили бы свежесть данных.

## 12. Как воспроизвести

1. Поднять стек и инициализировать БД:

```bash
make init-db DEPLOY_ENV=dev
docker compose -f deployments/dev/compose.yaml --env-file deployments/dev/.env up -d
```

2. Сгенерировать набор данных:

```bash
./perf_test/generate_data.sh dev
```

3. Снять baseline без оптимизаций:

```bash
docker exec -i -e PGPASSWORD=vkino dev-db-1 \
  psql -U vkino_user -d vkino < migrations/00008_perf_optimizations.down.sql
```

4. Прогнать `EXPLAIN ANALYZE` для запросов из разделов выше.

5. Вернуть оптимизации:

```bash
docker exec -i -e PGPASSWORD=vkino dev-db-1 \
  psql -U vkino_user -d vkino < migrations/00008_perf_optimizations.up.sql
```

6. Снова прогнать те же `EXPLAIN ANALYZE`.

7. Подготовить авторизованного пользователя для HTTP-нагрузки:

```bash
./perf_test/setup_load_user.sh dev
```

8. Установить `wrk` и гонять сценарии:

```bash
brew install wrk
./perf_test/run_wrk.sh selection-all
./perf_test/run_wrk.sh selection-one
./perf_test/run_wrk.sh movie-by-id
./perf_test/run_wrk.sh search-movie
./perf_test/run_wrk.sh favorites
./perf_test/run_wrk.sh watch-continue
./perf_test/run_wrk.sh watch-history
```

9. Очистить тестовые данные:

```bash
docker exec -i -e PGPASSWORD=vkino dev-db-1 \
  psql -U vkino_user -d vkino < perf_test/sql/cleanup_test_data.sql
```
