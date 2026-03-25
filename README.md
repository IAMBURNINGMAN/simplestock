# SimpleStock — Система складского учёта

Веб-приложение для управления складом: приход/расход товаров, инвентаризация, аналитика.

## Стек

**Backend:** Go, Chi, pgx, PostgreSQL, golang-migrate, bcrypt
**Frontend:** React 18, TypeScript, Vite, TailwindCSS, TanStack Query/Table, Recharts
**Инфра:** Docker Compose

## Функциональность

- Справочник товаров — CRUD, категории, SKU, единицы измерения, мин. остаток
- Приходные/расходные документы — создание, проведение, удаление черновиков
- Автоматический учёт остатков при проведении документов
- Инвентаризация — сверка факт/учёт, автокоррекция остатков
- История движений — полный лог операций с фильтрацией
- Дашборд — сводка по складу, алерты по низким остаткам
- Роли — admin / storekeeper

## Запуск

### Требования

- Docker и Docker Compose
- Go 1.25+
- Node.js 24+

### БД

```bash
docker compose up -d db
```

### Backend

```bash
cd backend
go run ./cmd/server
```

Сервер на `http://localhost:8080`. Миграции применяются автоматически.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Фронтенд на `http://localhost:5173`. Vite проксирует `/api/*` на бэкенд.

### Логин по умолчанию

`admin` / `admin`
