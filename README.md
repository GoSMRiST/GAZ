# GAZ Platform

Микросервисная платформа для организации встреч и активностей. Пользователи регистрируются, создают комнаты по интересам, вступают в них и общаются.

## Сервисы

| Сервис | Описание | Репозиторий |
|--------|----------|-------------|
| **auth_users** | Аутентификация, JWT, управление профилем | [GAZ_auth_users](https://github.com/GoSMRiST/GAZ_auth_users) |
| **rooms** | Создание и поиск комнат, участники, подписки | [GAZ_rooms](https://github.com/GoSMRiST/GAZ_rooms) |
| **ProtosGaz** | Общие protobuf-контракты для gRPC | [protosGaz](https://github.com/GoSMRiST/protosGaz) |

## Архитектура

```
┌─────────────────────────────────────────────────────┐
│                      Клиент                         │
└──────┬──────────────────────┬───────────────────────┘
       │                      │
       ▼                      ▼
┌─────────────┐       ┌──────────────┐
│  auth_users │       │    rooms     │
│             │◄─gRPC─│              │
│  REST :8080 │       │  REST :8083  │
│  REST :8081 │       └──────────────┘
│ gRPC :50051 │
│ gRPC :50052 │
└──────┬──────┘
       │
  ┌────┴────┐
  │  Redis  │  PostgreSQL
  └─────────┘  (users, rooms)
```

**Поток аутентификации:**
1. Клиент получает JWT через `POST /login` в `auth_users`
2. Все защищённые запросы отправляются с заголовком `Authorization: Bearer <token>`
3. Сервис `rooms` валидирует токен через gRPC-вызов к `auth_users` — без обращения к БД

## Технологии

- **Go 1.25** — все сервисы
- **PostgreSQL** — хранилище данных (отдельная БД на каждый сервис)
- **Redis** — rate limiting и защита от брутфорса
- **gRPC** + **Protocol Buffers** — межсервисное взаимодействие
- **Gin** — HTTP-роутинг
- **Docker** + **Docker Compose** — оркестрация

## Быстрый старт

### Требования

- Docker Desktop
- Docker Compose v2

### Запуск

```bash
git clone https://github.com/GoSMRiST/GAZ.git
cd GAZ

docker compose up --build
```

Сервисы поднимутся в правильном порядке автоматически:
1. PostgreSQL и Redis
2. Миграции (`migrate_auth`, `migrate_rooms`)
3. `auth_users`
4. `rooms`

### Порты

| Сервис | Адрес | Назначение |
|--------|-------|------------|
| auth_users | `localhost:8080` | Регистрация, вход, верификация email, сброс пароля |
| auth_users | `localhost:8081` | Профиль пользователя |
| rooms | `localhost:8083` | Комнаты |
| PostgreSQL | `localhost:5432` | — |
| Redis | `localhost:6379` | — |

## API

### auth_users — Auth REST `:8080`

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/register` | Регистрация |
| `POST` | `/login` | Вход → JWT |
| `GET` | `/verify?token=` | Подтверждение email |
| `POST` | `/forgot-password` | Запрос сброса пароля |
| `POST` | `/reset-password` | Смена пароля |

### auth_users — User REST `:8081`

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/user/me` | Свой профиль |
| `GET` | `/user/:id` | Профиль по ID |
| `PATCH` | `/nickname` | Обновить никнейм |
| `PATCH` | `/avatar` | Обновить аватар |
| `PATCH` | `/password` | Сменить пароль |
| `PATCH` | `/bio` | Обновить bio |

### rooms `:8083`

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| `GET` | `/room-types` | — | Типы комнат |
| `GET` | `/cities` | — | Города |
| `GET` | `/rooms` | — | Список комнат |
| `GET` | `/rooms/:id` | — | Комната по ID |
| `GET` | `/rooms/:id/participants` | — | Участники |
| `POST` | `/rooms` | ✓ | Создать комнату |
| `PUT` | `/rooms/:id` | ✓ | Обновить комнату |
| `DELETE` | `/rooms/:id` | ✓ | Удалить комнату |
| `POST` | `/rooms/:id/join` | ✓ | Вступить |
| `POST` | `/rooms/:id/leave` | ✓ | Покинуть |
| `GET` | `/users/me/rooms` | ✓ | Мои комнаты |

## Конфигурация

Каждый сервис читает конфигурацию из переменных окружения. При локальной разработке значения берутся из `internal/config/config.env`. При запуске через Docker Compose значения передаются через `environment:` в `docker-compose.yml`.

Подробнее — в README каждого сервиса:
- [auth_users/README.md](./auth_users/README.md)
- [rooms/README.md](./rooms/README.md)

## Структура репозитория

```
GAZ/
├── auth_users/          # сервис аутентификации
├── rooms/               # сервис комнат
├── static/              # статические файлы (аватары)
│   └── avatars/
├── docker/
│   └── init-db.sql      # создание баз данных при первом старте
└── docker-compose.yml   # оркестрация всех сервисов
```
