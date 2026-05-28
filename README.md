# agri-api

API REST pentru managementul mașinilor agricole, operatorilor și asignărilor, cu sistem complet de autentificare bazat pe JWT.

## Tehnologii

| Tehnologie | Rol |
|---|---|
| [Go](https://go.dev/) | Limbaj principal |
| [Gin](https://github.com/gin-gonic/gin) | Framework HTTP |
| [PostgreSQL](https://www.postgresql.org/) | Bază de date principală |
| [Redis](https://redis.io/) | Blacklist JWT (invalidare imediată a token-urilor) |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Migrații DB |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | Generare și validare JWT |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Hashing parole (cost 14) |
| [go-redis/v9](https://github.com/redis/go-redis) | Client Redis |
| [air](https://github.com/air-verse/air) | Live reload în development |
| [zap](https://github.com/uber-go/zap) | Logger structurat |

---

## Structura proiectului

```
agri-api/
├── cmd/
│   └── api/
│       └── main.go              # Punctul de intrare, dependency wiring
├── internal/
│   ├── auth/                    # Logica de autentificare
│   │   ├── jwt_service.go       # Generare/parsare JWT cu claim jti
│   │   ├── blacklist.go         # Blacklist token-uri în Redis
│   │   ├── refresh_service.go   # Gestionare refresh + reset tokens în DB
│   │   └── password.go          # Hash/verify parole cu bcrypt
│   ├── config/
│   │   └── config.go            # Încărcare configurație din .env
│   ├── db/
│   │   └── postgres.go          # Inițializare conexiune PostgreSQL
│   ├── redis/
│   │   └── client.go            # Inițializare conexiune Redis
│   ├── delivery/
│   │   └── http/
│   │       ├── router.go        # Înregistrare rute + AppDeps
│   │       ├── middleware/
│   │       │   └── auth.go      # JWT middleware cu verificare blacklist
│   │       └── handlers/
│   │           ├── auth.go      # Login, Register, Refresh, Logout, ForgotPassword, ResetPassword
│   │           ├── health.go
│   │           ├── machine_handler.go
│   │           ├── operator_handler.go
│   │           ├── assigment_handler.go
│   │           └── validation.go # Mesaje de eroare prietenoase pentru validator
│   ├── domain/                  # Modele de date pure
│   │   ├── user.go
│   │   ├── machine.go
│   │   ├── operator.go
│   │   └── assigment.go
│   ├── dto/                     # Transfer objects (paginare, răspunsuri)
│   ├── logger/
│   │   └── logger.go
│   ├── repository/              # Interfețe pentru accesul la date
│   │   └── postgres/            # Implementări concrete PostgreSQL
│   ├── store/                   # Agregator de repository-uri
│   └── usecase/                 # Logica de business
│       ├── auth_service.go
│       ├── machine_service.go
│       ├── operator_service.go
│       └── assigment_service.go
├── migrations/                  # Fișiere SQL versionare (up + down)
├── .env                         # Variabile de mediu (nu se comite)
├── .env.example                 # Exemplu de configurație
├── docker-compose.yml           # PostgreSQL + Redis în Docker
├── Makefile
└── go.mod
```

---

## Configurare

```bash
cp .env.example .env
```

Variabile necesare:

```env
APP_ENV=development

SERVER_HOST=0.0.0.0
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=agri
DB_PASSWORD=agri123
DB_NAME=agri_db
DB_SSLMODE=disable

JWT_SECRET=super-secret-key
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=7d

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=agri123
REDIS_DB=0

LOG_LEVEL=debug
```

---

## Pornire

```bash
# 1. Pornește PostgreSQL + Redis
docker compose up -d

# 2. Aplică migrațiile
make migrate-up

# 3. Pornește serverul (cu live reload)
make run
```

---

## Comenzi Makefile

| Comandă | Descriere |
|---|---|
| `make run` | Pornește serverul cu `air` (live reload) |
| `make migrate-up` | Aplică toate migrațiile noi |
| `make migrate-down` | Rollback ultima migrație |
| `make migrate-status` | Afișează versiunea curentă a migrațiilor |
| `make migrate-create name=nume` | Creează fișiere `.up.sql` și `.down.sql` noi |
| `make lint` | Rulează linter-ul |

---

## Endpoints

### Autentificare (publice)

| Metodă | Rută | Descriere |
|---|---|---|
| POST | `/api/auth/register` | Înregistrare utilizator nou |
| POST | `/api/auth/login` | Autentificare, returnează access + refresh token |
| POST | `/api/auth/refresh` | Rotație refresh token, blacklistare access token vechi |
| POST | `/api/auth/forgot-password` | Generare token de resetare parolă |
| POST | `/api/auth/reset-password` | Resetare parolă cu token |

### Autentificare (protejate — necesită `Authorization: Bearer <token>`)

| Metodă | Rută | Descriere |
|---|---|---|
| POST | `/api/logout` | Revocare refresh token + blacklistare access token |

### Health

| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/health` | Starea serviciului |

### Mașini (protejate)

| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/machines` | Listare mașini (paginare) |
| GET | `/api/machines/:id` | Obținere mașină după ID |
| POST | `/api/machines` | Creare mașină |
| PATCH | `/api/machines/:id` | Actualizare mașină |
| DELETE | `/api/machines/:id` | Ștergere mașină |

### Operatori (protejate)

| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/operators` | Listare operatori (paginare) |
| GET | `/api/operators/:id` | Obținere operator după ID |
| POST | `/api/operators` | Creare operator |
| PATCH | `/api/operators/:id` | Actualizare operator |
| DELETE | `/api/operators/:id` | Ștergere operator |

### Asignări (protejate)

| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/assignments` | Listare asignări (paginare) |
| GET | `/api/assignments/:id` | Obținere asignare după ID |
| POST | `/api/assignments` | Creare asignare |
| PATCH | `/api/assignments/:id` | Actualizare asignare |
| DELETE | `/api/assignments/:id` | Ștergere asignare |
| PATCH | `/api/assignments/:id/close` | Închidere asignare activă |

---

## Fluxul de autentificare

### Token-uri

| Token | TTL | Stocat în |
|---|---|---|
| Access token (JWT) | 15 minute | Client (header `Authorization`) |
| Refresh token (UUID) | 7 zile | DB `refresh_tokens` + client |

Fiecare access token conține un claim `jti` (UUID unic). La invalidare, JTI-ul e scris în Redis cu TTL egal cu timpul rămas al token-ului.

### Invalidare imediată

| Eveniment | Ce se întâmplă |
|---|---|
| **Login nou** | Toate sesiunile active sunt revocate din DB, JTI-urile lor sunt blacklistate în Redis |
| **Refresh** | JTI-ul access token-ului vechi (citit din DB) e blacklistat în Redis |
| **Logout** | Refresh token revocat în DB, access token blacklistat în Redis |
| **Request cu token revocat** | Middleware verifică Redis → `401 token has been revoked` |

---

## Arhitectură

```
HTTP Request
    ↓
middleware/auth.go       ← validare JWT + verificare blacklist Redis
    ↓
delivery/http/handlers/  ← validare input, răspuns HTTP
    ↓
usecase/                 ← logică de business
    ↓
repository/ (interfețe)  ← contract pentru date
    ↓
repository/postgres/     ← implementare SQL
    ↓
PostgreSQL / Redis
```

