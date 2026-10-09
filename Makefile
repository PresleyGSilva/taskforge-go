.PHONY: build up down test seed

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

test:
	cd backend && go test ./...
	cd frontend && npm test --if-present

seed:
	cd backend && go run ./cmd/seed
