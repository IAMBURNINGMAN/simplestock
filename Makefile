.PHONY: dev backend-dev frontend-dev db migrate seed build up down

# Development
dev: db backend-dev

db:
	docker compose up -d db

backend-dev:
	cd backend && go run ./cmd/server

frontend-dev:
	cd frontend && npm run dev



# Docker
up:
	docker compose up --build -d

down:
	docker compose down

# Build
build:
	cd backend && go build -o bin/server ./cmd/server
