# Smart One-Stop Digital Hearing & Automated Escalation Platform
### የ አንድ ማዕከል አገልግሎት እና አውቶማቲክ የቅሬታ መፍቻ ስርዓት

An end-to-end, production-grade Software Engineering Architecture implemented in **Go (Golang)** following **Clean Architecture** and **Domain-Driven Design (DDD)**. Designed for municipal administrations (e.g. Addis Ababa City Administration, Sub-Cities, and Woredas) to eliminate physical queuing, enforce Service Level Agreements (SLAs), and digitize citizen dispute resolution.

---

## 1. Architectural Topology & Clean Architecture Layers

```
       ┌────────────────────────────────────────────────────────────┐
       │                   Domain Layer (Core)                      │
       │   - Pure Business Rules (SLA Policies, Escalation Domain)   │
       │   - Domain Entities (Case, HearingSlot, Structure, User)   │
       │   - Repository & Port Interfaces (CaseRepository, etc.)    │
       └─────────────────────────────▲──────────────────────────────┘
                                     │
       ┌─────────────────────────────┴──────────────────────────────┐
       │                 Application / Use Cases                    │
       │   - Commands & Queries (Intake, Transition, Resolve)       │
       │   - Services: CaseService, SLAService, HearingService      │
       │   - DTOs: CaseResponse, AnalyticsSummary, HearingSlotDTO   │
       └─────────────────────────────▲──────────────────────────────┘
                                     │
       ┌─────────────────────────────┴──────────────────────────────┐
       │                  Infrastructure Layer                      │
       │   - PostgreSQL 16 (GORM driver + connection pooling)       │
       │   - SQLite (Local zero-dependency fallback driver)         │
       │   - Redis 7 & Distributed Locks (Redlock style)            │
       │   - MinIO / AWS S3 Evidence Storage Adapter                │
       │   - JWT / Bcrypt Cryptographic Security                    │
       └─────────────────────────────▲──────────────────────────────┘
                                     │
       ┌─────────────────────────────┴──────────────────────────────┐
       │                 Presentation / API Layer                   │
       │   - Gin HTTP Handlers & Middleware (RBAC, Tracing, CORS)   │
       │   - OpenAPI 3.0 / Swagger Documentation                    │
       │   - Modern Interactive Bilingual Web Portal (EN / አማርኛ)  │
       └────────────────────────────────────────────────────────────┘
```

---

## 2. Core Functional Capabilities

1. **Citizen Multi-Channel Intake & QR Verification:**
   - Citizens file grievances online selecting their Sub-City and Woreda.
   - Generates an instant cryptographic ticket (e.g., `TKT-2026-100234`) and QR verification code (`ETH-MUNI-...`).
   - Computes deterministic SLA deadlines based on service classifications (e.g. 24h, 48h, 72h).
2. **Race-Condition-Free Automated SLA Escalation Engine:**
   - Uses PostgreSQL `SELECT ... FOR UPDATE SKIP LOCKED` and Redis distributed locks.
   - If an official does not resolve a ticket within the mandated SLA, the engine automatically escalates the ticket to the next administrative echelon:
     $$\text{Woreda} \longrightarrow \text{Sub-City (ክፍለ ከተማ)} \longrightarrow \text{City Administration (ከተማ አስተዳደር)}$$
   - Increments escalation counts, automatically grants an SLA extension (+24h), and generates an append-only audit trail with the system actor.
3. **Digital Hearing Desk (Wednesday & Friday Desks):**
   - Strictly validates municipal calendar rules requiring hearing sessions to fall on **Wednesdays** or **Fridays** (`EXTRACT(ISODOW FROM hearing_date) IN (3, 5)`).
   - Generates and embeds WebRTC / Jitsi virtual meeting rooms.
   - Direct in-app resolution signing with certified digital stamps.
4. **Executive Observability & Bottleneck Heatmap:**
   - Real-time comparison of resolution efficiency and breach rates across municipal subdivisions.
   - Append-only cryptographic audit ledger tracking every state change.

---

## 3. Directory Layout

```
├── cmd/
│   ├── api/
│   │   └── main.go               # Primary API Server & embedded SLA daemon
│   └── worker/
│       └── main.go               # Dedicated background SLA Escalation Worker daemon
├── internal/
│   ├── domain/                   # Pure business entities, enums, errors, and port interfaces
│   │   ├── entity.go
│   │   ├── enums.go
│   │   ├── errors.go
│   │   └── repository.go
│   ├── application/              # Use cases, DTOs, and domain services
│   │   ├── dto/
│   │   │   └── dto.go
│   │   └── service/
│   │       ├── case_service.go
│   │       ├── sla_service.go
│   │       ├── hearing_service.go
│   │       ├── auth_service.go
│   │       └── analytics_service.go
│   ├── infrastructure/           # Database, Redis, Security, and concrete SQL adapters
│   │   ├── config/config.go
│   │   ├── database/
│   │   │   ├── database.go
│   │   │   └── seeder.go
│   │   ├── redis/client.go
│   │   ├── security/
│   │   │   ├── jwt.go
│   │   │   └── hasher.go
│   │   └── repository/
│   │       ├── case_repo.go
│   │       ├── structure_repo.go
│   │       ├── user_repo.go
│   │       ├── citizen_repo.go
│   │       ├── hearing_repo.go
│   │       ├── audit_repo.go
│   │       └── service_type_repo.go
│   └── presentation/             # Gin HTTP handlers, declarative RBAC, and routing
│       ├── handler/
│       │   ├── auth_handler.go
│       │   ├── case_handler.go
│       │   ├── hearing_handler.go
│       │   ├── structure_handler.go
│       │   └── analytics_handler.go
│       ├── middleware/
│       │   ├── auth_middleware.go
│       │   ├── correlation.go
│       │   ├── idempotency.go
│       │   ├── cors.go
│       │   └── logger.go
│       └── router.go
├── web/
│   └── static/                   # Modern responsive Bilingual Web Portal (EN / አማርኛ)
│       ├── index.html
│       ├── css/style.css
│       └── js/app.js
├── Dockerfile                    # Multi-stage production container
├── docker-compose.yml            # Multi-service enterprise stack (Postgres, Redis, MinIO, API, Worker)
├── openapi.yaml                  # Full OpenAPI 3.0 specification
└── go.mod
```

---

## 4. Quickstart Guide

### Option A: Direct Local Execution (Zero External Dependencies)

The platform supports an automatic SQLite and in-process cron fallback. You can run the application directly with Go:

```bash
# 1. Run the API Server
go run cmd/api/main.go
```

Open your browser to:
👉 **http://localhost:8080**

- The API server will automatically initialize SQLite, run auto-migrations, and seed municipal data.
- The embedded SLA cron daemon runs every 60 seconds.

---

### Option B: Full Enterprise Stack (Docker Compose)

To spin up the complete distributed microservice-ready environment (PostgreSQL 16, Redis 7, MinIO S3, API server, and dedicated SLA worker):

```bash
# Start all containers
docker compose up --build -d

# Verify health status
docker compose ps
```

Services exposed:
- **Web Portal & REST API:** `http://localhost:8080`
- **PostgreSQL 16:** `localhost:5432` (`postgres` / `postgres123`)
- **Redis 7:** `localhost:6379`
- **MinIO Console:** `http://localhost:9001` (`admin` / `minioadmin123`)

---

## 5. Pre-Seeded Default Accounts & Demonstration Data

The seeder initializes the municipal hierarchy for Addis Ababa:
- **City:** Addis Ababa City Administration
- **Sub-Cities:** Kirkos Sub-City, Bole Sub-City
- **Woredas:** Kirkos Woreda 01, Kirkos Woreda 02, Bole Woreda 03

### Staff Logins (Password for all: `Password123!`):

| Role | Name | Email | Jurisdiction |
| :--- | :--- | :--- | :--- |
| **SUPER_ADMIN** | Alemayehu Tadesse | `superadmin@smartonestop.gov.et` | City Administration |
| **CITY_DIRECTOR** | Dr. Selamawit Bekele | `city.director@smartonestop.gov.et` | City Administration |
| **SUBCITY_MANAGER** | Kassahun Haile | `kirkos.manager@smartonestop.gov.et` | Kirkos Sub-City |
| **WOREDA_OFFICER** | Bethlehem Girma | `woreda01.officer@smartonestop.gov.et` | Kirkos Woreda 01 |
| **SERVICE_DESK_AGENT** | Yonas Mulugeta | `desk.agent@smartonestop.gov.et` | Kirkos Woreda 01 |

### Pre-Seeded Demonstration Tickets:

- `TKT-2026-100234`: Active boundary dispute under review at Woreda 01 (SLA countdown green).
- `TKT-2026-100582`: Commercial license grievance **already breached**! Click **"Run SLA Escalation Engine"** on the dashboard to watch it automatically escalate from Woreda 01 to Kirkos Sub-City!
- `TKT-2026-100891`: Kebele house tenancy transfer scheduled for Wednesday digital hearing desk.

---

## 6. Ethiopian Working-Day Engine & Civic Calendar

The platform features an enterprise **Ethiopian Working-Day & Holiday Engine** implemented in `internal/domain/workcalendar` with 0 external I/O and ~95% test statement coverage:

- **13-Month Ethiopian Calendar:** Algorithmic conversion using Julian Day Number (JDN) mapping between Gregorian and Ethiopic epochs (`ethiopicJDNEpoch = 1723856`).
- **Leap Year Dynamics:** Computes 6 days in Pagume (ጳጉሜ) every 4 years (`year % 4 == 3`), automatically handling the +1 day Gregorian holiday shift following Ethiopian leap years.
- **Orthodox Ecclesiastical Computus:** Determines movable Orthodox Good Friday (*Siklet*) and Easter (*Fasika*) using the Julian-calendar Dionysian paschalion computus with 13-day Gregorian adjustment.
- **Working-Day SLA Counting:** Automatically calculates SLA deadlines and remaining hours by skipping weekends and official public holidays.
- **Digital Hearing Scheduler:** Guarantees hearings are scheduled strictly on working **Wednesdays and Fridays**, automatically skipping public holidays.
- **Ge'ez Numerals & Bilingual Names:** Converts dates and numbers to authentic Ge'ez numerals (፩, ፪, ፲, ፳, ፻) and displays month names in English and Amharic (መስከረም ... ጳጉሜ).

### Calendar API Endpoints:
- `GET /api/v1/calendar/working-days?start=2026-10-01&hours=24` - Calculate deadline and working days
- `GET /api/v1/calendar/holidays?year=2026` - List all seeded public holidays (fixed, movable, Islamic)
- `GET /api/v1/calendar/next-hearing-slots?from=2026-10-01&count=5` - Next available Wednesday/Friday hearing sessions

---

## 7. Responsive Mobile-First Redesign & Collapsible Navigation

The portal implements an accessible, responsive civic design system across 5 viewports:
- **Breakpoints:** `360px` (compact mobile), `480px` (large mobile), `768px` (tablets/drawers), `1024px` (laptops), and `1440px+` (command displays).
- **Responsive Table Transformation:** On screens `< 768px`, data tables transform into elevated card lists with `data-label` metadata and 2x2 touch action grids, eliminating horizontal scrolling.
- **Mobile Bottom Sheets:** Modals and case dossiers transform into mobile slide-up sheets with top touch handles, fluid scrolling, and sticky action buttons.
- **Collapsible Staff Sidebar:**
  - **Expanded (260px):** Brand header, route indicators, pending case counter badge, officer profile card, and SLA trigger button.
  - **Collapsed Rail (72px):** Icon-only presentation with hover tooltips and smooth cubic-bezier transition.
  - **Mobile Off-Canvas Drawer (`< 768px`):** Slides out with background backdrop overlay, hamburger button in top command bar, swipe-to-close, and keyboard shortcut (`Ctrl+B` / `Cmd+B`).
- **Touch-First Accessibility:** Minimum 44x44px touch targets across all interactive elements.

---

## 8. Testing, Quality Assurance & CI

### Running Unit & Package Tests
```bash
# Run all tests with coverage
go test -v -cover ./...

# Run internal/domain/workcalendar test suite with breakdown
go test -v -coverprofile=coverage.out ./internal/domain/workcalendar
go tool cover -func=coverage.out

# Run Go Vet
go vet ./...

# Validate Client-Side Vanilla JS
node -c web/static/js/app.js web/static/js/calendar.js
```

### Continuous Integration (CI)
GitHub Actions workflow configured in `.github/workflows/ci.yml` validates:
1. `go vet ./...` and `go build ./...`
2. `go test -v -race -cover ./...`
3. Node syntax check (`node -c`) for zero frontend build step integrity
4. Docker Compose configuration verification (`docker compose config`)

---

## 9. Key API Endpoints Reference

- `GET /healthz` - Liveness probe
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/docs` - OpenAPI 3.0 specification
- `POST /api/v1/cases` - Citizen grievance intake (`X-Idempotency-Key` supported)
- `GET /api/v1/cases/ticket/{ticket}` - Public ticket status & timeline query
- `GET /api/v1/cases/verify/{qr}` - Public QR cryptographic verification
- `POST /api/v1/cases/{id}/resolve` - Sign off official resolution & certified directive
- `POST /api/v1/cases/{id}/escalate` - Manual supervisor escalation
- `POST /api/v1/cases/{id}/hearings` - Schedule Wednesday/Friday digital hearing slot
- `GET /api/v1/analytics/dashboard` - Executive metrics & bottleneck heatmap
- `POST /api/v1/sla/trigger-sweep` - On-demand trigger for SLA auto-escalation sweep
- `GET /api/v1/calendar/working-days` - Working-day calculation & deadline prediction
- `GET /api/v1/calendar/holidays` - Public holidays list
- `GET /api/v1/calendar/next-hearing-slots` - Available hearing appointment slots

