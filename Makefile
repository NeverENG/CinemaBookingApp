.PHONY: run migrate test vet compose-up db-reset

run:
	go run ./cmd/lterm

migrate:
	go run ./cmd/lterm -migrate

test:
	go test ./...

vet:
	go vet ./...

compose-up:
	docker compose up --build

db-reset:
	dropdb --if-exists cinema
	createdb cinema
	go run ./cmd/lterm -migrate
