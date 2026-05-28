include .env
export

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATE=migrate -path ./migrations -database "$(DB_URL)"

.PHONY: run migrate-up migrate-down migrate-status migrate-create fmt lint swagger

## Pornește serverul cu live-reload
run:
	air

## Aplică toate migrațiile noi
migrate-up:
	$(MIGRATE) up

## Rollback ultima migrație
migrate-down:
	$(MIGRATE) down 1

## Afișează statusul migrațiilor
migrate-status:
	$(MIGRATE) version

## Creează o migrație nouă: make migrate-create name=nume_migratie
migrate-create:
	migrate create -ext sql -dir ./migrations -seq $(name)

## Formatează tot codul Go
fmt:
	go fmt ./...

## Rulează linter-ul
lint:
	golangci-lint run ./...

## Generează documentația Swagger
swagger:
	swag init -g cmd/api/main.go --output docs
