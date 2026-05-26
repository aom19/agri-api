# agri-api

API REST pentru managementul mașinilor agricole, operatorilor și asignărilor dintre aceștia.

## Tehnologii

| Tehnologie | Rol |
|---|---|
| [Go 1.25](https://go.dev/) | Limbaj principal |
| [Gin](https://github.com/gin-gonic/gin) | Framework HTTP |
| [PostgreSQL](https://www.postgresql.org/) | Bază de date |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Migrații DB |
| [air](https://github.com/air-verse/air) | Live reload în development |
| [zap](https://github.com/uber-go/zap) | Logger structurat |

---

## Structura proiectului

```
agri-api/
├── cmd/
│   └── api/
│       └── main.go              # Punctul de intrare al aplicației
├── internal/
│   ├── config/
│   │   └── config.go            # Încărcare configurație din .env
│   ├── db/
│   │   └── postgres.go          # Inițializare conexiune PostgreSQL
│   ├── delivery/
│   │   └── http/
│   │       ├── router.go        # Înregistrare rute + AppDeps
│   │       └── handlers/        # Un handler per resursă
│   │           ├── health.go
│   │           ├── machine_handler.go
│   │           ├── operator_handler.go
│   │           └── assigment_handler.go
│   ├── domain/                  # Modele de date pure, fără dependențe externe
│   │   ├── machine.go
│   │   ├── operator.go
│   │   └── assigment.go
│   ├── logger/
│   │   └── logger.go            # Wrapper peste zap
│   ├── repository/              # Interfețele pentru accesul la date
│   │   ├── machine_repository.go
│   │   ├── operator_repository.go
│   │   ├── assigment_repository.go
│   │   └── postgres/            # Implementările concrete pentru PostgreSQL
│   │       ├── machine_repo.go
│   │       ├── operator_repo.go
│   │       └── assigment_repo.go
│   └── usecase/                 # Logica de business
│       ├── machine_service.go
│       ├── operator_service.go
│       └── assigment_service.go
├── migrations/                  # Fișiere SQL de migrație (up + down)
├── .env                         # Variabile de mediu (nu se comite)
├── .env.example                 # Exemplu de configurație
├── .air.toml                    # Configurație live reload
├── docker-compose.yml           # PostgreSQL în Docker
├── Makefile                     # Comenzi utile
└── go.mod
```

### Motivarea structurii de foldere

**`cmd/api/`** — Convenție Go standard ([golang-standards/project-layout](https://github.com/golang-standards/project-layout)). `cmd/` poate găzdui mai multe binare (ex: `cmd/worker/`, `cmd/migrate/`). Fiecare subfolder devine un executabil separat.

**`internal/`** — Pachet privat Go: codul din `internal/` nu poate fi importat de module externe. Protejează implementarea internă și impune granițe clare.

**`internal/domain/`** — Conține structurile de date centrale (`Machine`, `Operator`, `Assigment`) și constantele de status. Nu importă nimic din proiect — este nucleul pur al aplicației (inspirat din *Clean Architecture*).

**`internal/delivery/http/`** — Stratul de livrare HTTP. Separat de logică pentru că în viitor poate fi adăugat și un alt tip de livrare (gRPC, CLI). `handlers/` conține câte un fișier per resursă pentru claritate.

**`internal/repository/`** — Interfețele definesc *contractul* pentru accesul la date. `postgres/` conține implementările concrete. Această separare permite înlocuirea PostgreSQL cu altă bază de date fără să se modifice logica de business.

**`internal/usecase/`** — Logica de business pură. Serviciile nu știu despre HTTP sau SQL — comunică doar prin interfețele din `repository/`. Aceasta este convenția *usecase layer* din Clean Architecture.

**`migrations/`** — Fișiere SQL versionare cu format `000001_nume.up.sql` / `000001_nume.down.sql` compatibil cu `golang-migrate`. Fiecare migrație are și varianta de rollback (`down`).

---

## Configurare

Copiază `.env.example` și completează valorile:

```bash
cp .env.example .env
```

Variabile necesare:

```env
APP_ENV=development          # development | production

SERVER_HOST=0.0.0.0
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=agri
DB_PASSWORD=agri123
DB_NAME=agri_db
DB_SSLMODE=disable
```

---

## Pornire

### 1. Pornește baza de date

```bash
docker compose up -d
```

### 2. Rulează migrațiile

```bash
make migrate-up
```

### 3. Pornește serverul (cu live reload)

```bash
make run
# sau direct: air
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

---

## Endpoints

### Health
| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/health` | Starea serviciului |

### Mașini
| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/machines` | Listare toate mașinile |
| GET | `/api/machines/:id` | Obținere mașină după ID |
| POST | `/api/machines` | Creare mașină nouă |
| PATCH | `/api/machines/:id` | Actualizare mașină |

### Operatori
| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/operators` | Listare toți operatorii |
| GET | `/api/operators/:id` | Obținere operator după ID |
| POST | `/api/operators` | Creare operator nou |
| PUT | `/api/operators/:id` | Actualizare operator |
| DELETE | `/api/operators/:id` | Ștergere operator |

### Asignări
| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/assignments` | Listare toate asignările |
| GET | `/api/assignments/:id` | Obținere asignare după ID |
| POST | `/api/assignments` | Creare asignare (validează că mașina și operatorul sunt liberi) |
| PATCH | `/api/assignments/:id` | Actualizare asignare |
| DELETE | `/api/assignments/:id` | Ștergere asignare (doar dacă nu e activă) |
| POST | `/api/assignments/:id/close` | Închidere asignare activă |

---

## Arhitectură

Proiectul urmează principiile **Clean Architecture**:

```
HTTP Request
    ↓
delivery/http/handlers   ← validare input, răspuns HTTP
    ↓
usecase/                 ← logică de business, validări, tranzacții
    ↓
repository/ (interfețe) ← contract pentru date
    ↓
repository/postgres/     ← implementare SQL
    ↓
PostgreSQL
```

Fiecare strat depinde doar de stratul de dedesubt prin **interfețe**, nu prin implementări concrete. Aceasta permite testarea fiecărui strat în izolare.

### Dependency Injection

Dependențele sunt injectate manual în `main.go` și grupate în `AppDeps`:

```go
httpdelivery.SetupRoutes(server, httpdelivery.AppDeps{
    Log:              log,
    MachineService:   machineService,
    OperatorService:  operatorService,
    AssigmentService: assigmentService,
})
```

Adăugarea unui serviciu nou necesită doar o linie nouă în `AppDeps` — semnătura `SetupRoutes` nu se modifică.
