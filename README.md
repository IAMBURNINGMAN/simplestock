# SimpleStock — Система складского учёта

Веб-приложение для управления складом: приход/расход товаров, инвентаризация, аналитика и отчёты.

## Стек

**Backend:** Go 1.25, Chi router, pgx (PostgreSQL), golang-migrate, bcrypt-сессии
**Frontend:** React 18, TypeScript, Vite, TailwindCSS, TanStack Query/Table, Recharts
**БД:** PostgreSQL 16
**Инфра:** Docker Compose

## Структура проекта

```
backend/
├── cmd/server/          — точка входа
├── internal/
│   ├── config/          — конфигурация из env
│   ├── domain/          — доменные модели
│   ├── dto/             — request/response структуры
│   ├── handler/         — HTTP-хэндлеры
│   ├── middleware/       — авторизация, сессии
│   ├── repository/      — работа с БД (pgx)
│   └── service/         — бизнес-логика
├── migrations/          — SQL-миграции
└── Dockerfile

frontend/
├── src/
│   ├── api/             — HTTP-клиент
│   ├── components/      — Layout, UI-компоненты
│   ├── hooks/           — useAuth и кастомные хуки
│   └── pages/           — страницы приложения
└── vite.config.ts
```

## Функциональность

- **Справочник товаров** — CRUD, категории, SKU, единицы измерения, мин. остаток
- **Приходные/расходные документы** — создание, проведение, отмена
- **Автоматический учёт остатков** — при проведении документа quantity обновляется
- **Инвентаризация** — сверка факт/учёт, автокоррекция остатков
- **История движений** — лог всех операций по товарам
- **Дашборд** — сводка по складу, графики (Recharts)
- **Отчёты** — аналитика по движениям и остаткам
- **Роли** — admin / storekeeper

## Запуск

### Требования

- Docker и Docker Compose
- Go 1.25+ (для локальной разработки бэкенда)
- Node.js 24+ (для локальной разработки фронтенда)

### Через Docker (полный стек)

```bash
docker compose up -d
```

Backend доступен на `http://localhost:8080`, фронтенд собирается и раздаётся бэкендом.

### Локальная разработка

1. Поднять БД:
```bash
docker compose up -d db
```

2. Запустить бэкенд:
```bash
cd backend
go run ./cmd/server
```

3. Запустить фронтенд:
```bash
cd frontend
npm install
npm run dev
```

Фронтенд — `http://localhost:5173`, бэкенд — `http://localhost:8080`.
Vite проксирует `/api/*` на бэкенд автоматически.

### Переменные окружения (backend)

| Переменная     | По умолчанию                                              | Описание           |
|----------------|-----------------------------------------------------------|--------------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/simplestock`  | Строка подключения |
| `SESSION_KEY`  | `simplestock-super-secret-key`                            | Ключ сессии        |
| `PORT`         | `8080`                                                    | Порт сервера       |

### Вход по умолчанию

Логин: `admin` / Пароль: `admin`

## Схема БД

- `users` — пользователи (admin, storekeeper)
- `categories` — категории товаров
- `products` — товары (SKU, остаток, мин. остаток, цена)
- `documents` — приходные/расходные документы (draft → completed)
- `document_items` — позиции документа
- `movements` — история движений товаров
- `inventories` — инвентаризации
- `inventory_items` — позиции инвентаризации (факт vs учёт)
