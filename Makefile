APP      := lterm
DB_DSN   ?= host=localhost user=4ge0 dbname=cinema port=5432 sslmode=disable TimeZone=Asia/Shanghai
TEST_DB  ?= host=localhost user=4ge0 dbname=cinema_test port=5432 sslmode=disable TimeZone=Asia/Shanghai

.PHONY: run migrate build test test-integration fmt vet compose-up compose-down

run:
	DB_DSN="$(DB_DSN)" go run ./cmd/lterm

migrate:
	DB_DSN="$(DB_DSN)" go run ./cmd/lterm -migrate

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
