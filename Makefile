.PHONY: run migrate test test-integration vet compose-up db-reset

run:
	go run ./cmd/lterm

migrate:
	go run ./cmd/lterm -migrate

test:
	go test ./...

test-integration:
	TEST_DB_DSN=$${TEST_DB_DSN:?set TEST_DB_DSN} go test ./internal/infra/database/postgres/ -run Integration -v

vet:
	go vet ./...

compose-up:
	docker compose up --build

db-reset:
	dropdb --if-exists cinema
	createdb cinema
	go run ./cmd/lterm -migrate
