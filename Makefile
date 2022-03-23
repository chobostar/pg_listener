.PHONY: up
up:
	docker-compose up -d

.PHONY: down
down:
	docker-compose kill && docker-compose rm

.PHONY: init
init:
	docker-compose exec -T postgres psql -U postgres < ./demo/db/init.sql
	docker-compose exec -T postgres psql -U postgres < ./demo/db/slot.sql

.PHONY: build
build:
	go build -o ./bin/pg_listener ./cmd/pg_listener.go

.PHONY: test
test:
	go test -v ./...

.PHONY: demo
demo:
	PGLSN_DB_HOST=localhost \
	PGLSN_DB_PORT=5432 \
	PGLSN_DB_USER=postgres \
	PGLSN_DB_PASS=postgres \
	PGLSN_DB_NAME=postgres \
	PGLSN_TABLE_NAMES=queue.events \
	PGLSN_SLOT=pg_listener \
	PGLSN_KAFKA_HOSTS=localhost:9092 \
	PGLSN_CHUNKS=1 \
	go run ./cmd/pg_listener.go -debug

