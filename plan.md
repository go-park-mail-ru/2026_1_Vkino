# План по уменьшению дублирования в `pkg/errmap/*`

## Что дублируется сейчас

По коду видно два уровня дублирования:

1. Внутри самих мапперов:
   - [`pkg/errmap/httpx/mapper.go`](/home/saucesamba/Рабочий%20стол/vkino-back/pkg/errmap/httpx/mapper.go)
   - [`pkg/errmap/grpcx/mapper.go`](/home/saucesamba/Рабочий%20стол/vkino-back/pkg/errmap/grpcx/mapper.go)
   - [`pkg/errmap/mapper.go`](/home/saucesamba/Рабочий%20стол/vkino-back/pkg/errmap/mapper.go)

   И `httpx`, и `grpcx` хранят:
   - `order`
   - `rules`
   - `default`
   - цикл поиска правила

   Хотя generic-механика уже есть в `pkg/errmap/mapper.go`.

2. В описании правил для gRPC:
   - [`services/auth-service/internal/delivery/grpc/errors.go`](/home/saucesamba/Рабочий%20стол/vkino-back/services/auth-service/internal/delivery/grpc/errors.go)
   - [`services/user-service/internal/delivery/grpc/errors.go`](/home/saucesamba/Рабочий%20стол/vkino-back/services/user-service/internal/delivery/grpc/errors.go)
   - [`services/movie-service/internal/delivery/grpc/errors.go`](/home/saucesamba/Рабочий%20стол/vkino-back/services/movie-service/internal/delivery/grpc/errors.go)

   Там повторяется один и тот же паттерн:
   - много ошибок в `order`
   - map с одинаковыми `codes.Code`
   - одинаковые тексты (`unauthorized`, `internal server error`, `not found`, и т.д.)
   - дублирование domain error + repo error с одной и той же реакцией

## Что я бы предложил

### Вариант 1. Минимальный рефакторинг

Оставить публичный API `httpx.New(...)` и `grpcx.New(...)`, но внутри перевести их на generic mapper из `pkg/errmap/mapper.go`.

### Идея

- `pkg/errmap/mapper.go` оставить как ядро
- `pkg/errmap/grpcx/mapper.go` сделать thin wrapper над generic mapper
- `pkg/errmap/httpx/mapper.go` сделать thin wrapper над generic mapper

### Что получится

- логика поиска правила будет в одном месте
- `Map()` в `grpcx` и `httpx` останется как адаптер результата
- внешний код сервисов почти не изменится

### Плюсы

- минимальный риск
- почти без миграции по проекту
- резко уменьшает дублирование инфраструктурного кода

### Минусы

- дублирование самих наборов правил в сервисах останется

### Когда выбирать

Если цель сейчас: убрать архитектурный мусор без большой переделки.

---

### Вариант 2. Нормализовать описание правил через `RuleSet`

Вынести описание правил в общий формат и сделать builder/helper для `grpcx` и `httpx`.

### Идея

Вместо такого стиля:

```go
grpcx.New(
    []error{...},
    map[error]grpcx.ErrResponse{...},
    codes.Internal,
    "internal server error",
)
```

перейти к чему-то вроде:

```go
errmap.NewErrorSet(
    errmap.Rule(domain.ErrUserNotFound, codes.NotFound, "user not found"),
    errmap.Rule(postgresrepo.ErrUserNotFound, codes.NotFound, "user not found"),
    errmap.Rule(domain.ErrInvalidToken, codes.Unauthenticated, "unauthorized"),
)
```

а дальше:

```go
grpcMapper := grpcx.NewFromSet(set, codes.Internal, "internal server error")
```

или

```go
httpMapper := httpx.NewFromCodeSet(...)
```

### Что получится

- описания правил станут компактнее
- исчезнет двойная структура `order + map`
- порядок можно сохранить прямо в slice правил

### Плюсы

- код сервисов станет короче и легче читать
- исчезнет постоянное дублирование `order` и `map`
- легче добавлять новые ошибки

### Минусы

- это уже API-рефакторинг
- потребуется изменить все текущие места создания мапперов

### Когда выбирать

Если цель: сделать систему маппинга заметно чище, а не только убрать внутреннее дублирование.

---

### Вариант 3. Ввести групповые helper'ы для типовых ошибок

Поверх варианта 2 добавить helper-функции для частых шаблонов.

### Идея

Сейчас много повторов такого типа:

- `ErrInvalidToken`, `ErrPasswordMismatch`, `ErrInvalidCredentials` -> `Unauthenticated`, `"unauthorized"`
- `repo.ErrUserNotFound`, `domain.ErrUserNotFound` -> `NotFound`, `"user not found"`
- `domain.ErrInternal` -> `Internal`, `"internal server error"`

Можно сделать helper'ы:

```go
errmapgrpc.Unauthorized(
    domain.ErrInvalidToken,
    domain.ErrInvalidCredentials,
    domain.ErrPasswordMismatch,
)

errmapgrpc.NotFound("user not found",
    domain.ErrUserNotFound,
    postgresrepo.ErrUserNotFound,
)
```

### Что получится

- самые длинные файлы `errors.go` станут в 2-3 раза короче
- намерение будет читаться лучше, чем список однотипных записей

### Плюсы

- сильнее всего убирает визуальный шум
- хорошо масштабируется на новые сервисы

### Минусы

- это уже более opinionated abstraction
- можно переусложнить, если helper'ов станет слишком много

### Когда выбирать

Если тебе важна именно компактность и единообразие service-level error mapping.

## Что я рекомендую

Оптимальный путь для этого проекта:

| Этап | Что сделать | Зачем |
|---|---|---|
| 1 | Перевести `httpx` и `grpcx` на generic `pkg/errmap/mapper.go` | убрать инфраструктурное дублирование без риска |
| 2 | Заменить `order + map` на единый `[]Rule` | убрать лишнюю сложность в описании правил |
| 3 | Добавить 2-4 helper'а для частых шаблонов (`Unauthorized`, `NotFound`, `InvalidArgument`, `Internal`) | уменьшить дублирование в `services/*/delivery/grpc/errors.go` |

Это даст хороший баланс:

- код станет короче
- публичная модель будет проще
- не будет ощущения "магии ради магии"

## Как бы я это реализовал

### Шаг 1. Сделать общую сущность правила

Примерно так:

```go
type Rule[K comparable, R any] struct {
    Key      K
    Response R
}
```

И generic mapper должен принимать не `order + map`, а:

```go
[]Rule[K, R]
```

Почему это лучше:

- порядок уже зашит в slice
- не надо синхронизировать `order` и `rules`
- меньше шансов ошибиться

### Шаг 2. Переписать generic mapper под `[]Rule`

Тогда ядро станет примерно таким:

```go
type Mapper[S any, K comparable, R any] struct {
    rules []Rule[K, R]
    match Matcher[S, K]
}
```

### Шаг 3. Сделать адаптеры для `grpcx` и `httpx`

#### Для `grpcx`

- subject: `error`
- key: `error`
- match: `errors.Is`
- response: `{Code, Message}`

#### Для `httpx`

- subject: `codes.Code`
- key: `codes.Code`
- response: `{Status, Message}`

И лучше мапить не `error -> status` напрямую, а в два шага:

1. `error -> gRPC status`
2. `gRPC code -> HTTP response`

Тогда `httpx` становится по сути таблицей преобразования `codes.Code -> HTTP`
и не дублирует идею error-matching вообще.

Это особенно важно: сейчас `httpx` и `grpcx` выглядят как две разные системы,
а по смыслу `httpx` у тебя уже работает поверх gRPC статусов.

## Самое практичное упрощение

Если коротко: я бы делал так.

### Для `grpcx`

Перевести все `errors.go` на `[]Rule`, например:

```go
var authGRPCErrorMapper = grpcx.New(
    []grpcx.Rule{
        {Err: domain.ErrUserAlreadyExists, Code: codes.AlreadyExists, Message: "user already exists"},
        {Err: postgresrepo.ErrUserAlreadyExists, Code: codes.AlreadyExists, Message: "user already exists"},
        {Err: domain.ErrInvalidCredentials, Code: codes.Unauthenticated, Message: "unauthorized"},
    },
    codes.Internal,
    "internal server error",
)
```

### Для `httpx`

Сделать вообще плоскую таблицу:

```go
var DefaultMapper = httpx.New(
    []httpx.Rule{
        {Code: codes.InvalidArgument, Status: http.StatusBadRequest},
        {Code: codes.NotFound, Status: http.StatusNotFound},
        {Code: codes.AlreadyExists, Status: http.StatusConflict},
    },
    http.StatusInternalServerError,
    "internal server error",
)
```

Тогда `httpx` станет очень простым, потому что ему не нужен `errors.Is`,
не нужен отдельный generic matching по error chain, и не нужна сложная структура.

## Вывод

Если цель именно "убрать лютое дублирование", то лучший план такой:

1. Убрать `order + map` и перейти на `[]Rule`.
2. Использовать `pkg/errmap/mapper.go` как единое ядро, а не держать две почти одинаковые реализации рядом.
3. Для service-level gRPC мапперов ввести helper'ы для повторяющихся шаблонов.
4. Для `httpx` оставить очень плоскую таблицу `gRPC code -> HTTP status/message`, без лишней абстракции.

Если захочешь, следующим сообщением я могу уже предложить не просто идею, а конкретный целевой API: как именно должны выглядеть новые `Rule`, `New`, helper'ы и один полностью переписанный пример `errors.go`.
