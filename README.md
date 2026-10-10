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
| [Open-Meteo](https://open-meteo.com/) | Provider gratuit pentru vremea curentă din Cantemir |

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
│   │   ├── user.go              # User, UserProfile
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
│       ├── profile_service.go   # GetProfile, UpdateProfile, UploadPhoto
│       ├── machine_service.go
│       ├── operator_service.go
│       └── assigment_service.go
├── migrations/                  # Fișiere SQL versionare (up + down)
├── uploads/
│   └── avatars/                 # Poze de profil încărcate (excluse din git)
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

PUBLIC_URL=http://localhost:8080
OPENWEATHER_API_KEY=your_openweathermap_api_key

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
| `make test` | Rulează testele unitare |
| `make test-cover` | Rulează testele și generează `coverage.out` + `test-report.json` (pentru SonarQube) |
| `make test-db` | Rulează testele de integrare pe Postgres real, într-o bază `<DB_NAME>_test` recreată la fiecare rulare |

---

## Teste

Testele sunt unitare și rulează fără Postgres, Redis sau SMTP:

```bash
make test          # toate testele
make test-cover    # + coverage.out și test-report.json, citite de SonarQube
go test ./internal/usecase/test/ -run TestMachineService   # un singur serviciu
```

**Unde stau:** fiecare pachet are testele într-un subfolder `test/` (ex. `internal/usecase/test/`), în pachetul extern `<pachet>_test`. Testele văd deci doar ce e exportat. Funcțiile și câmpurile interne pe care le verifică sunt expuse în `export_for_tests.go` din pachetul testat (`config`, `handlers`, `usecase`). Aplicația nu folosește nimic din aceste fișiere.

**Cum sunt scrise:**
- Serviciile din `internal/usecase/` primesc mock-uri ale repository-urilor (`internal/usecase/test/mocks_test.go`): fiecare mock are câmpuri-funcție, iar testul setează doar metodele de care are nevoie.
- Serviciile care deschid tranzacții (`*sql.DB`) folosesc [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) pentru `Begin`/`Commit`/`Rollback` și pentru query-urile din `internal/auth/refresh_service.go`.
- Redis și SMTP sunt înlocuite cu adrese inaccesibile (`127.0.0.1:1`): blacklist-ul și e-mailurile sunt best-effort, deci testele verifică că eroarea e tratată, nu că mesajul a ajuns.
- Handler-ele HTTP se testează cu `httptest` și `gin` peste servicii construite pe mock-uri (`internal/delivery/http/handlers/test/`); `internal/delivery/http/test/router_test.go` verifică tabela de rutare.

### Teste de integrare (Postgres real)

SQL-ul din `internal/repository/postgres` se testează pe o bază reală, în `internal/repository/postgres/test/`:

```bash
make docker-infra   # dacă Postgres nu rulează deja
make test-db
```

- Baza `<DB_NAME>_test` (implicit `agri_db_test`) e ștearsă și recreată din `migrations/*.up.sql` la fiecare rulare, deci testele verifică și că migrațiile merg pe o bază nouă. Baza de dezvoltare nu e atinsă: testele refuză orice bază al cărei nume nu se termină în `_test`.
- Fiecare test golește tabelele de date și își inserează singur datele, direct în SQL.
- Testele trec prin serviciile reale, nu doar prin repository: mișcările de stoc (inclusiv ieșiri simultane, care verifică blocarea pe rând), recolta în stoc (doar diferența, corecții refuzate fără urme, cereri simultane) și rapoartele Flotă și Operatori.
- Fără `TEST_DATABASE_URL`, testele sunt sărite, deci `make test` rămâne fără dependențe externe.

**Ce nu intră în coverage** (vezi `sonar.coverage.exclusions`): `cmd/` (pornirea aplicației), `internal/db`, `internal/redis` și `internal/repository/postgres` (SQL care are nevoie de o bază reală; e acoperit de `make test-db`, care nu intră în raportul Sonar). Restul e măsurat, inclusiv handler-ele HTTP, care au deocamdată doar câteva teste; adăugarea de teste pentru restul handler-elor e cel mai simplu mod de a crește procentul.

> `make test` rulează doar pachetele care au fișiere de test. Motivul: toolchain-ul Go 1.25 descărcat automat (`GOTOOLCHAIN=auto`) nu include unealta `covdata`, de care `go test -cover` are nevoie pentru pachetele fără teste. Pachetele fără teste apar oricum în `coverage.out` cu 0%, prin `-coverpkg=./...`.
>
> Pentru coverage folosește `make test-cover`, nu `go test -cover`. Testele stau în alt pachet decât codul testat, deci fără `-coverpkg` procentul afișat e al pachetului de test, nu al codului.

---

## Analiză de cod (SonarQube)

SonarQube Community (26.9) rulează local în `docker-compose.yml` și analizează atât API-ul (Go, Dockerfile, șabloane HTML), cât și frontend-ul (TypeScript). Serviciul e în profilul `sonar`, deci **nu** pornește cu `docker compose up` / `make docker-infra`, pentru că ocupă ~2 GB RAM.

**Prima configurare (o singură dată):**

```bash
make sonar-up      # pornește SonarQube și așteaptă să fie gata (~45s)
# deschide http://localhost:9000, intră cu admin / admin și schimbă parola
make sonar-token   # generează un token de analiză și îl salvează în .env (SONAR_TOKEN)
```

**Utilizare:**

| Comandă | Descriere |
|---|---|
| `make sonar-up` | Pornește SonarQube (după un restart de Docker) |
| `make sonar` | Analizează API-ul și frontend-ul |
| `make sonar-api` | Analizează doar API-ul |
| `make sonar-front` | Analizează doar frontend-ul (din `../agri-front`) |
| `make sonar-down` | Oprește SonarQube; datele rămân în volume |

Rezultatele se văd la http://localhost:9000. Scanner-ul rulează din Docker, deci nu trebuie instalat nimic.

**De știut:**
- `SONAR_TOKEN` e personal pentru fiecare instalare și stă doar în `.env` (necomis). Se regenerează cu `make sonar-token` dacă ștergi volumele SonarQube.
- Configurarea analizei e în `sonar-project.properties`. Excepțiile de reguli se pun tot acolo, cu motivul scris în comentariu, nu în interfața SonarQube: fiecare dezvoltator are propriul server local, deci ce marchezi în UI rămâne doar la tine.
- Excepție existentă: regulile despre atribute HTML învechite și tabele de layout sunt dezactivate pentru `internal/email/templates/`, pentru că Outlook și mulți clienți de e-mail nu suportă layout CSS.
- `make sonar-api` rulează întâi `make test-cover`, iar `make sonar-front` rulează `npm run test:coverage` în `../agri-front`, ca SonarQube să primească și coverage-ul. Ce e exclus din procent e listat (cu motiv) în `sonar-project.properties`.

---

## Endpoints

### Profil utilizator (protejate)

| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/profile` | Obținere profil utilizator autentificat (incl. email, rol, foto) |
| PATCH | `/api/profile` | Actualizare first_name, last_name, date_of_birth |
| POST | `/api/profile/photo` | Încărcare poză de profil (multipart/form-data, câmp `photo`, max 5MB, jpg/png/webp) |

**Fișiere statice:** poza de profil e servită la `/uploads/avatars/<filename>`.

### Meteo (public)

| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/weather/current` | Vremea curentă pentru Cantemir, normalizată și cache-uită în backend |
| GET | `/api/weather/current?lat={lat}&lng={lng}&location={name}` | Vremea curentă pentru coordonate specifice, folosită pentru meteo per teren |

Backendul folosește `OPENWEATHER_API_KEY` pentru OpenWeatherMap când cheia este configurată. Dacă providerul nu răspunde sau cheia este respinsă, endpointul cade automat pe Open-Meteo, care nu necesită API key pentru uz non-comercial sub limita publică de cereri. Backendul face request-ul extern, aplică timeout și cache de 10 minute, iar frontendul consumă doar endpointul intern. Endpointul este public deoarece nu expune date sensibile.

Răspunsul include metrici utile pentru hartă și agricultură: temperatură, condiție meteo, umiditate, vânt, precipitații pe ultima oră și nebulozitate. Cheia OpenWeatherMap rămâne doar în backend.

### Profil utilizator (protejate)

| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/profile` | Obținere profil utilizator autentificat (incl. email, rol, foto) |
| PATCH | `/api/profile` | Actualizare first_name, last_name, date_of_birth |
| POST | `/api/profile/photo` | Încărcare poză de profil (multipart/form-data, câmp `photo`, max 5MB, jpg/png/webp) |

**Fișiere statice:** poza de profil e servită la `/uploads/avatars/<filename>`.

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

### Dashboard (protejate)

| Metodă | Rută | Descriere |
|---|---|---|
| GET | `/api/dashboard/cards` | Carduri KPI agregate din baza de date: total mașini, mașini active, total operatori și alocări active din operațiunile pe teren |

Necesită permisiunea `dashboard:read`, acordată implicit rolurilor `admin`, `manager` și `viewer` prin migrația `000030_add_dashboard_read_permission`. Migrațiile `000032_add_operator_dashboard_permissions` și `000033_limit_operator_permissions` creează/asigură rolul `operator` și îi limitează accesul la `dashboard:read` și `field_operations:read`; operațiunile pe teren sunt filtrate în backend după operatorul asociat userului curent.

Răspunsurile pentru `/api/field-operations` și `/api/field-operations/:id` includ `field_geometry` (GeoJSON Polygon), `machine_status` și `implement_status`, astfel încât operatorii pot vedea conturul terenului și disponibilitatea resurselor lucrării fără permisiuni separate pentru modulele de terenuri, mașini sau echipamente.

Pornirea lucrării se face prin `/api/field-operations/:id/start`, care verifică că mașina și echipamentul asignate sunt `active` înainte să schimbe statusul în `in_progress`. Checklistul de plecare cu 4 bife salvate a fost scos (migrația `000049_drop_field_operation_checklist`); în pagină rămâne o singură confirmare, „Am verificat utilajul și terenul”, care nu se salvează.

Regula de compatibilitate are un singur nivel: template-ul operațiunii poate limita tipurile de mașini și de echipamente (`template_machine_types`, `template_implement_types`). La crearea și editarea unei operațiuni pe teren, API-ul răspunde cu 400 dacă mașina sau echipamentul are alt tip decât cele din template; un template fără tipuri, sau o operațiune fără template, acceptă orice. Tabelul `implement_compatibilities` și tipurile de mașini permise operatorilor au fost șterse (migrația `000050_drop_compatibility_duplicates`).

### Audit și notificări (protejate)

Migrația `000036_create_audit_and_notifications` creează jurnalul de audit, notificările și permisiunile `audit:read` și `notifications:read`. Jurnalul este disponibil pentru admin la `/api/audit-log` și poate fi filtrat după `entity_type`, `entity_id` și `limit`; frontendul îl afișează în pagina `/admin/audit`.

Handler-ele pentru mașini, echipamente, resurse, mișcări de stoc, terenuri, operatori, alocări, utilizatori, template-uri și operațiuni pe teren primesc serviciile de audit/notificări din router. Actualizările de status pentru mașini/echipamente/operatori/alocări, nivelurile minime de stoc și pornirea lucrărilor emit notificări persistente și evenimente WebSocket pe `/ws/notifications?token=<jwt>`. Operațiunile de creare, actualizare, ștergere și start sunt logate în audit cu `actor_id` din tokenul JWT; răspunsul audit include `actor_name` din profilul utilizatorului și `entity_name` pentru denumirea entității afectate, iar schimbările de status salvează tranziția `old_status` -> `status` când statusul anterior este disponibil.

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

