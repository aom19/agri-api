include .env
export

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATE=migrate -path ./migrations -database "$(DB_URL)"

.PHONY: run migrate-up migrate-down migrate-status migrate-create fmt lint swagger \
        docker-infra docker-dev docker-prod docker-build seed \
        sonar-up sonar-down sonar-token sonar sonar-api sonar-front sonar-check

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

## ─── Docker ───────────────────────────────────────────────────────────────────

## Pornește doar infrastructura locală (postgres + redis) — pentru dev cu air
docker-infra:
	docker compose up -d

## Oprește infrastructura locală
docker-infra-down:
	docker compose down

## Stack complet dev (app + infra) cu build
docker-dev:
	docker compose -f infrastructure/compose/docker-compose.dev.yml up --build

## Stack producție (detached)
docker-prod:
	docker compose -f infrastructure/compose/docker-compose.prod.yml up -d --build

## Construiește imaginea Docker a aplicației
docker-build:
	docker build -t agri-api:latest .

## ─── Seed ────────────────────────────────────────────────────────────────────

## Populează baza de date cu date de test (rulează manual, nu la migrate-up)
seed:
	docker exec -i agri_postgres psql -U $(DB_USER) -d $(DB_NAME) < seeds/seed.sql

## Populează roluri, permisiuni și relațiile RBAC (rulează după migrate-up)
seed-rbac:
	docker exec -i agri_postgres psql -U $(DB_USER) -d $(DB_NAME) < seeds/rbac_seed.sql

## ─── SonarQube ───────────────────────────────────────────────────────────────

# Scanner-ul rulează în rețeaua containerului SonarQube, deci vede serverul pe localhost:9000
# (funcționează la fel pe Docker Desktop și pe Linux nativ).
SONAR_URL=http://localhost:9000
# $(call sonar_scan,<director proiect>) — directorul trebuie să conțină sonar-project.properties.
sonar_scan=docker run --rm --network container:agri_sonarqube \
	-e SONAR_HOST_URL=$(SONAR_URL) -e SONAR_TOKEN=$(SONAR_TOKEN) \
	-v "$(1):/usr/src" sonarsource/sonar-scanner-cli:latest

## Pornește SonarQube și așteaptă să fie gata (http://localhost:9000)
sonar-up:
	docker compose --profile sonar up -d --wait sonarqube

## Oprește SonarQube (datele rămân în volume)
sonar-down:
	docker compose --profile sonar stop sonarqube

## Generează un token de analiză și îl scrie în .env (o singură dată; cere parola de admin)
sonar-token:
	@read -p "Utilizator SonarQube [admin]: " user; user=$${user:-admin}; \
	read -s -p "Parola: " pass; echo; \
	token=$$(curl -fs -u "$$user:$$pass" -X POST \
		"$(SONAR_URL)/api/user_tokens/generate?name=local-scanner-$$(date +%s)&type=GLOBAL_ANALYSIS_TOKEN" \
		| sed -n 's/.*"token":"\([^"]*\)".*/\1/p'); \
	if [ -z "$$token" ]; then echo "Nu am putut genera token-ul (parolă greșită sau SonarQube oprit?)"; exit 1; fi; \
	sed -i -e '/^SONAR_TOKEN=/d' -e '$$a\' .env; echo "SONAR_TOKEN=$$token" >> .env; \
	echo "Token salvat în .env"

## Analizează API-ul și frontend-ul
sonar: sonar-api sonar-front

## Analizează doar API-ul (Go)
sonar-api: sonar-check
	$(call sonar_scan,$(CURDIR))

## Analizează doar frontend-ul (TypeScript)
sonar-front: sonar-check
	$(call sonar_scan,$(abspath $(CURDIR)/../agri-front))

sonar-check:
	@if [ -z "$(SONAR_TOKEN)" ]; then echo "Lipsește SONAR_TOKEN în .env — rulează: make sonar-token"; exit 1; fi
	@if [ "$$(docker inspect -f '{{.State.Health.Status}}' agri_sonarqube 2>/dev/null)" != "healthy" ]; then \
		echo "SonarQube nu rulează — pornește-l cu: make sonar-up"; exit 1; fi
