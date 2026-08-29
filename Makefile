APP      := lterm
DB_DSN   ?= host=localhost user=4ge0 dbname=cinema port=5432 sslmode=disable TimeZone=Asia/Shanghai
TEST_DB  ?= host=localhost user=4ge0 dbname=cinema_test port=5432 sslmode=disable TimeZone=Asia/Shanghai

.PHONY: run migrate seed build test test-integration fmt vet compose-up compose-down compose-logs compose-ps compose-reset smoke

run:
	DB_DSN="$(DB_DSN)" go run ./cmd/lterm

migrate:
	DB_DSN="$(DB_DSN)" go run ./cmd/lterm -migrate -seed

seed:
	DB_DSN="$(DB_DSN)" go run ./cmd/lterm -seed

build:
	go build -o bin/$(APP) ./cmd/lterm

test:
	go test ./...

test-integration:
	TEST_DB_DSN="$(TEST_DB)" go test ./internal/infra/database/postgres -v

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f migrate backend frontend

compose-ps:
	docker compose ps

compose-reset:
	docker compose down -v --remove-orphans

smoke:
	./scripts/smoke.sh http://127.0.0.1:$${APP_PORT:-8088}
