# FILAS — Go/Gin + Vue 3 modernization

A full **technology modernization** of the FILAS website (Fundación para la
Integración Laboral y Autonomía Social) — the same application, rebuilt on a
modern stack.

> [!IMPORTANT]
> **This is a reversion of the _technologies_, not of the _look_.**
> The user interface is deliberately preserved: every view was ported
> faithfully from the legacy pages — same layout, same branding, same Spanish
> copy. What changed is everything _underneath_: the plain-PHP API became a
> Go/Gin service with Clean Architecture, the vanilla-JS multi-page site became
> a Vue 3 SPA, and the whole thing runs on Docker Compose. A visitor sees the
> same site; a developer opens a completely modern codebase.

The original project was a school/portfolio app: a plain-PHP API (no framework,
a hand-rolled router, business logic living in MySQL triggers) served to a
multi-page vanilla-JavaScript front end. This repository migrates it while
keeping the **same MySQL schema and the same API contract**, so behavior is
preserved (quirks and all — see [`backend/docs/legacy-quirks.md`](backend/docs/legacy-quirks.md)).

It is a **local-only portfolio project**: production-like polish, zero
production operations (no TLS, no secrets manager, no scaling).

---

## Stack: before → after

| Layer     | Legacy                                   | Now                                                        |
| --------- | ---------------------------------------- | --------------------------------------------------------- |
| Backend   | Plain PHP, no framework, `global $conn`  | **Go 1.23 + Gin**, Clean Architecture, `sqlx`             |
| Auth      | JWT + `sha256(salt+password)`            | JWT + **bcrypt** with transparent migrate-on-login        |
| Business logic | MySQL **triggers**                  | Ported into the Go **usecase** layer; triggers dropped    |
| Frontend  | Vanilla JS, 10 HTML pages, hash router   | **Vue 3** (`<script setup>`) + Vite + vue-router + Pinia  |
| Styling   | Hand-compiled per-page SCSS/CSS          | Component-scoped SCSS with shared tokens                  |
| Data      | MySQL (unchanged schema)                 | **MySQL 8** in Docker, seeded from the same dump          |
| Delivery  | Manual / devtunnel                       | **Docker Compose** (one-command demo via nginx)           |
| Tests     | none                                      | Go (`testing` + sqlmock) & Vitest, wired into CI          |

---

## Architecture

Dependencies point **inward**: `handler → usecase → domain ← repository`. The
domain layer imports nothing; wiring happens only in the composition root
(`cmd/api/main.go`). This keeps a classic ports-and-adapters shape without
over-engineering a ~7-resource CRUD API.

```
┌──────────── Browser ────────────┐
│   Vue 3 SPA (vue-router, Pinia)  │
└───────────────┬─────────────────┘
                │  /api/*  (same-origin; nginx reverse-proxy in the demo profile,
                │           Vite dev-server proxy in development)
┌───────────────▼─────────────────┐
│         Go / Gin  (:8080)        │
│  handler (REST + DTO + middleware)
│      │                           │
│  usecase (business logic,        │
│      │    ported trigger logic)  │
│  domain (entities, interfaces)   │
│      ▲                           │
│  repository/mysql (sqlx)         │
└───────────────┬─────────────────┘
                │
        ┌───────▼────────┐
        │   MySQL 8       │
        └────────────────┘
```

**Backend layers**

| Layer        | Package                      | Responsibility                                                        |
| ------------ | --------------------------- | --------------------------------------------------------------------- |
| Domain       | `internal/domain`           | Entities, repository interfaces, sentinel errors, invariants          |
| Usecase      | `internal/usecase`          | Application logic, validation, the ported order-trigger logic, auth   |
| Repository   | `internal/repository/mysql` | `sqlx` implementations, parameterized SQL, transactional order writes  |
| Handler      | `internal/handler/rest`     | Gin handlers, request/response DTOs, HTTP status mapping, JWT middleware |

**Frontend**

A single-page app: `vue-router` (history mode) replaces the hand-rolled hash
router; `Pinia` stores hold auth/cart state; `api/client.js` is one fetch
interceptor (bearer-token injection + central 401 → logout); reusable
composables (`useAsyncResource`, `useResourceCrud`) remove per-view boilerplate.
Public views and an authenticated admin panel (per-resource CRUD) both consume
the Go API.

---

## Project structure

```
.
├── backend/                     # Go / Gin API
│   ├── cmd/api/main.go          # composition root (config, DB, wiring, server)
│   ├── internal/
│   │   ├── domain/              # entities + repository interfaces + errors
│   │   ├── usecase/             # business logic (incl. ported order triggers)
│   │   ├── repository/mysql/    # sqlx repositories (parameterized SQL)
│   │   ├── handler/rest/        # Gin handlers, dto/, middleware/
│   │   ├── auth/                # JWT + password (bcrypt + legacy sha256)
│   │   └── config/              # env → typed config
│   ├── db/init/01-schema.sql    # seed (schema + data, triggers dropped)
│   ├── docs/legacy-quirks.md    # characterization of the legacy contract
│   └── Dockerfile
├── frontend/                    # Vue 3 SPA
│   ├── src/
│   │   ├── api/                 # client.js (interceptor) + resources.js
│   │   ├── stores/              # Pinia: auth, cart
│   │   ├── composables/         # useAsyncResource, useResourceCrud
│   │   ├── views/public/        # Home, Noticias, Galería, Familias, …, WildeArtesanal
│   │   ├── views/admin/         # login, layout, per-resource CRUD
│   │   ├── components/          # NavBar, ProductCard, ResourceForm, …
│   │   └── styles/              # SCSS tokens + mixins (@use)
│   ├── public/assets/           # static images (served at /assets/*)
│   ├── Dockerfile               # node build → nginx serve
│   └── nginx.conf               # SPA fallback + /api reverse proxy
├── docker-compose.yml           # db + api  (+ web under the `demo` profile)
├── .env.example                 # copy to .env
├── filas.sql                    # original dump (kept for reference)
└── .github/workflows/           # backend-ci.yml + frontend-ci.yml
```

---

## Getting started

Requires Docker (with Compose). For the dev workflow you also need Node 20+ and
Go 1.23+.

```bash
cp .env.example .env
```

### Demo (one command)

Builds the SPA, serves it with nginx, and reverse-proxies `/api` to the Go
container — same-origin, no CORS:

```bash
docker compose --profile demo up --build
```

Then open **http://localhost:8081** (the `WEB_PORT` in `.env`).

Admin panel: **/admin** — local dev credentials are seeded and documented in
`backend/db/init/01-schema.sql` (header note 3). They are synthetic; there are
no real secrets in this repository.

### Development (hot reload)

Run the database and API in containers, and the Vite dev server on the host so
you get HMR. Vite proxies `/api` to the Go container:

```bash
docker compose up -d db api      # MySQL + Go API
cd frontend
npm install
npm run dev                      # http://localhost:5173
```

### Configuration

All configuration is environment-driven; copy `.env.example` → `.env` and
adjust if a port collides. Key variables: `DB_*` (database), `API_PORT`,
`WEB_PORT`, `JWT_SECRET`, `JWT_EXPIRY_HOURS`, `CORS_ALLOWED_ORIGINS`,
`VITE_API_URL`. The committed `.env.example` carries working local values —
this is a portfolio project, so there are no production secrets to protect.

---

## API overview

REST under `/api`, plus `/health`. Reads are public; writes require a JWT
(`Authorization: Bearer <token>`), except `POST /api/orders`, which is the
public customer checkout.

| Resource        | Routes                                             |
| --------------- | -------------------------------------------------- |
| `products`      | `GET` (public) · `POST/PUT/DELETE` (auth)          |
| `news`          | `GET` (public) · `POST/PUT/DELETE` (auth)          |
| `gallery`       | `GET` (public) · `POST/PUT/DELETE` (auth)          |
| `family`        | `GET` (public) · `POST/PUT/DELETE` (auth)          |
| `organizations` | `GET` (public) · `POST/PUT/DELETE` (auth)          |
| `orders`        | `POST` (public checkout) · `GET/PUT/PATCH` (auth)  |
| `admins`        | `POST {login:true}` (login) · CRUD (auth)          |

The Go DTOs reproduce the legacy JSON shapes byte-for-byte (including
mysqli's string-typed numeric fields and the orders' `GROUP_CONCAT` array),
so the contract is preserved. Intentional divergences are documented in
[`backend/docs/legacy-quirks.md`](backend/docs/legacy-quirks.md).

---

## Testing & CI

```bash
# backend
cd backend && go test ./...

# frontend
cd frontend && npm test          # Vitest
```

Two GitHub Actions workflows run on their respective paths:
`backend-ci.yml` (`go vet` + `go test`) and `frontend-ci.yml`
(`npm ci` + test + build).

---

## Engineering highlights

- **Strangler-fig, backend-first.** The PHP API was the contract oracle: Go
  responses were **characterized byte-for-byte** against the live legacy PHP on
  identical seeded data before each slice was accepted.
- **MySQL triggers ported to Go.** Five triggers (line price, order total,
  finish-date stamping, stock) moved into the order usecase, proven equivalent
  by a before/after gate, then dropped from the seed — the database no longer
  owns business logic.
- **Atomic, race-safe orders.** Order creation is one transaction; stock is
  decremented with a floor-guarded conditional `UPDATE` (a CAS), so concurrent
  checkouts cannot oversell into negative stock.
- **Transparent password upgrade.** Legacy `sha256(salt+password)` hashes are
  verified and then re-hashed to bcrypt on first successful login — no forced
  reset.
- **Security fixes as deliberate divergences.** e.g. the legacy news endpoints
  accepted unauthenticated writes; the Go backend enforces JWT on all writes.
- **Delivered in reviewable slices** under strict TDD, each passing an
  adversarial multi-lens review before being sealed.

---

## Scope & limitations (by design)

Local-only portfolio project — **not** production-ready, on purpose:

- No TLS, no secrets manager, no horizontal scaling.
- The committed seed keeps the original data as-is (including some legacy
  integrity garbage); forward invariants prevent new bad rows.
- Known, documented follow-ups (out of scope here): deadlock-retry on the
  order write under heavy same-product contention, `PUT`-order stock
  reconciliation, and exposing a per-line product id for the admin order
  editor.
