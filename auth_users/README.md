# auth_users

Микросервис аутентификации и управления пользователями платформы GAZ.

## Описание

Реализует полный цикл аутентификации: регистрация с подтверждением email, вход с защитой от брутфорса, выдача JWT-токенов, сброс пароля по ссылке, управление профилем. Предоставляет REST и gRPC интерфейсы — другие сервисы валидируют токены и получают данные пользователя через gRPC без обращения к базе данных.

## Архитектура

Сервис запускает четыре сервера одновременно:

| Сервер | Протокол | Назначение |
|--------|----------|------------|
| Auth REST | HTTP | Регистрация, вход, верификация email, сброс пароля |
| User REST | HTTP | Профиль пользователя, обновление данных |
| Auth gRPC | gRPC | Валидация JWT-токенов для других сервисов |
| User gRPC | gRPC | Получение данных пользователя (верификация, подписка) |

Все серверы завершают работу корректно (graceful shutdown) по сигналам `SIGTERM` / `SIGINT`.

## Технологии

- **Go 1.25** — структурированные логи через стандартный `log/slog`
- **Gin** — HTTP-роутинг
- **gRPC** — межсервисное взаимодействие
- **PostgreSQL** + **pgx/v5** — пул соединений, миграции через `golang-migrate`
- **Redis** — ограничение частоты входа и счётчики брутфорса
- **JWT** (golang-jwt/v5) — stateless токены доступа
- **bcrypt** — хэширование паролей
- **SMTP** — отправка писем верификации и сброса пароля
- **Фоновый воркер** — каждый час очищает просроченные записи `email_verifications` и `password_resets`

## Структура проекта

```
auth_users/
├── cmd/
│   ├── app/            # точка входа
│   └── migrator/       # запуск миграций
├── internal/
│   ├── app/            # сборка и запуск всех четырёх серверов
│   │   ├── grpcapp/    # запуск gRPC серверов
│   │   └── restapp/    # запуск REST серверов
│   ├── config/         # конфигурация через переменные окружения
│   ├── core/           # доменные ошибки и DTO
│   ├── infrastructure/
│   │   ├── email/      # SMTP клиент
│   │   └── redis/      # Redis клиент
│   ├── jwt/            # генерация и парсинг токенов
│   ├── middleware/      # middleware аутентификации
│   ├── repository/     # запросы к PostgreSQL
│   ├── service/
│   │   ├── auth/       # регистрация, вход, email-flow
│   │   ├── user/       # управление профилем
│   │   └── jwt/        # сервис валидации токенов
│   ├── transport/
│   │   ├── auth/       # REST + gRPC обработчики аутентификации
│   │   └── user/       # REST + gRPC обработчики пользователя
│   └── worker/         # фоновый воркер очистки
└── migrations/         # SQL миграции (up/down)
```

## API

### Auth REST

| Метод | Путь | Авторизация | Описание |
|-------|------|-------------|----------|
| `POST` | `/register` | — | Регистрация; отправляет письмо верификации |
| `POST` | `/login` | — | Вход; возвращает JWT |
| `GET` | `/verify?token=` | — | Подтверждение email |
| `POST` | `/forgot-password` | — | Запрос письма сброса пароля |
| `GET` | `/reset-password?token=` | — | HTML-форма смены пароля |
| `POST` | `/reset-password` | — | Отправка нового пароля |

### User REST

| Метод | Путь | Авторизация | Описание |
|-------|------|-------------|----------|
| `GET` | `/user/me` | Bearer JWT | Свой профиль |
| `GET` | `/user/:id` | Bearer JWT | Публичный профиль по ID |
| `PATCH` | `/nickname` | Bearer JWT | Обновить никнейм |
| `PATCH` | `/avatar` | Bearer JWT | Обновить URL аватара |
| `POST` | `/avatar-upload` | Bearer JWT | Загрузить аватар |
| `PATCH` | `/password` | Bearer JWT | Сменить пароль |
| `PATCH` | `/bio` | Bearer JWT | Обновить bio |

### gRPC

Контракты описаны в [ProtosGaz](https://github.com/GoSMRiST/protosGaz):

- **Token сервис** — `ValidateToken(token) → user_id` — валидация JWT без обращения к БД
- **User сервис** — `GetUserInfo(user_id) → verified, subscription` — проверка верификации и уровня подписки

## Поток аутентификации

```
POST /register → сохранить неверифицированного пользователя → отправить письмо с токеном
GET /verify?token= → создать пользователя → 200 OK
POST /login → проверить пароль (bcrypt) → проверить Redis rate limit → выдать JWT
JWT → другие сервисы вызывают gRPC ValidateToken → получают user_id
```

**Защита от брутфорса:** неудачные попытки входа считаются в Redis по email. После превышения порога аккаунт временно блокируется.

## Конфигурация

Заполни `internal/config/config.env`. В продакшене переменные передаются через окружение — файл `.env` игнорируется если переменные уже заданы.

| Переменная | Описание |
|------------|----------|
| `REST_AUTH_HOST_ADDRESS` | Адрес Auth REST сервера (например `:8080`) |
| `REST_USER_HOST_ADDRESS` | Адрес User REST сервера (например `:8081`) |
| `GRPC_AUTH_PORT` | Порт Auth gRPC (например `50051`) |
| `GRPC_USER_PORT` | Порт User gRPC (например `50052`) |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | Подключение к PostgreSQL |
| `REDIS_ADDRESS` / `REDIS_PASSWORD` / `REDIS_DB` | Подключение к Redis |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USER` / `SMTP_PASS` | Отправка писем |
| `APP_URL` | Базовый URL для ссылок в письмах |
| `JWT_SECRET` | HMAC-секрет для подписи токенов |
| `TOKEN_TTL` | Время жизни токена (например `24h`) |
| `SERV_TIMEOUT` | Таймаут graceful shutdown (по умолчанию `5s`) |
| `LOG_LEVEL` | `local` / `dev` / `prod` |

## Миграции

```bash
go run ./cmd/migrator
```

Миграции находятся в `migrations/`, выполняются по порядку. Каждая миграция имеет файлы `up` и `down`.

## Запуск локально

```bash
# 1. Запустить зависимости
docker run -d -p 5432:5432 -e POSTGRES_USER=user -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=users postgres:16
docker run -d -p 6379:6379 redis:7

# 2. Заполнить конфиг
# отредактировать internal/config/config.env

# 3. Применить миграции
go run ./cmd/migrator

# 4. Запустить
go run ./cmd/app
```

## Сборка

```bash
go build -o auth_users ./cmd/app
go build -o migrator ./cmd/migrator
```
