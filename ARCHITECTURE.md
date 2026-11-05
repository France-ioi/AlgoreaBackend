# AlgoreaBackend Architecture

**Targeted at agents.** Last updated: 2026-11-05.

## Overview

Go REST API for Algorea (France IOI): users/groups, educational items, results, permissions, attempts, help threads. Emphasizes permission/results propagation through hierarchies, multi-domain config, and external auth. Governed business rules and cross-service behavior: [docs/agents/domain.md](docs/agents/domain.md) (knowledge base `docs/algorea/`).

Stack: Go 1.26, MySQL 8, chi, GORM v1 (wrapped), godog BDD, viper, cobra, goose. Deploy: Docker and AWS Lambda (`algnhsa`). Details live in `go.mod` / `docker-compose.yml`.

## Deep-dives

Keep this file as the **map**. Put branch-specific detail under `docs/architecture/` and link it here. Read a deep-dive only when the change touches that branch:

| Branch | Doc |
|--------|-----|
| DataStore, transactions, retries, CTEs, query helpers | [database.md](docs/architecture/database.md) |
| API packages, handlers, errors, `service.Base` | [api-services.md](docs/architecture/api-services.md) |
| Sessions, permissions, domains | [auth.md](docs/architecture/auth.md) |
| Permissions/results propagation, observability | [propagation.md](docs/architecture/propagation.md) |
| Domain events, SQS, versioning | [events.md](docs/architecture/events.md) |
| Unit / integration / BDD, coverage recipe | [testing.md](docs/architecture/testing.md) |
| Viper config, propagation/async, session params | [configuration.md](docs/architecture/configuration.md) |
| CORS fail-closed defaults and credentials rules | [cors.md](docs/architecture/cors.md) |
| CLI commands (`serve`, `propagation`, `handle-event`, …) | [cli.md](docs/architecture/cli.md) |
| Core tables, `display_settings`, migrations vs `schema.sql` | [schema.md](docs/architecture/schema.md) |

When a change affects global structure (folders, request flow, shared patterns), update **this** file. When it only deepens one branch above, update that deep-dive.

## Project structure

```
AlgoreaBackend/
├── app/
│   ├── api/           # endpoints by resource (answers, auth, groups, items, …)
│   ├── auth/          # auth middleware
│   ├── database/      # DataStore, stores, DB wrappers
│   ├── domain/        # multi-domain config
│   ├── service/       # Base, AppHandler, errors, params
│   ├── token/         # JWT
│   ├── app.go         # init
│   └── server.go      # HTTP server
├── cmd/               # cobra commands
├── conf/              # config samples (active yml gitignored)
├── db/migrations/     # goose migrations
├── db/schema/         # base schema for db-restore
├── golang/            # generic helpers
├── testhelpers/       # godog steps
└── main.go
```

## Request flow

```
HTTP Request
  → Middleware (VersionHeader, Cache-Control: no-store, RealIP, RequestID, compress?, logger, Recoverer, CORS, Domain)
  → API router → service router (JSON + UserMiddleware)
  → Handler (params, user, store, logic)
  → AppHandler (panic/error → APIError → JSON)
  → Response
```

App lifecycle: `main.go` → cobra (`serve`, `db-*`, Lambda default) → `app.New()` (config, logger, DB, router) → `server.Start()`.

Middleware order and DataStore/User injection: [api-services.md](docs/architecture/api-services.md), [auth.md](docs/architecture/auth.md).

Every API response is non-cacheable (`Cache-Control: no-store`); a handler may override this.

## Conventions (every change)

- One handler per file; service embeds `service.Base`; routes in `SetRoutes()`
- Stores via `DataStore` factories; API errors via `service.ErrXxx()`; unexpected → `MustNotBeError`
- Writes in `InTransaction` / `EnsureTransaction`; schedule propagation inside the transaction when groups/items/permissions/results change
- Prefer query builders + `With()` CTEs; quote/escape via `database.QuoteName` / `EscapeLikeString`
- Context: logger, user, session, domain config (see auth/api deep-dives)
- Migrations: `YYMMDDHHMM_description.sql|.go`; never edit old migrations or hand-edit `db/schema/schema.sql` — [schema.md](docs/architecture/schema.md)
- New events: type + `Dispatch` after commit + BDD + [events.md](docs/architecture/events.md)
- Never disable linter rules; BDD for new API endpoints; check permissions before mutating
