# Configuration

## Loading

Priority: `Set()` → flags → env → file → defaults.

Files: `conf/config.yml` (skipped in test env), `conf/config.{env}.yml`. Env prefix `ALGOREA_`; nested keys with `__` (e.g. `ALGOREA_DATABASE__ADDR`).

```go
config := app.LoadConfig()
dbConfig, _ := app.DBConfig(config)
serverConfig := app.ServerConfig(config)
domainConfig, _ := app.DomainsConfig(config)
```

## Sections agents touch most

**Database** — `sessionParams` re-applied after every pool checkout (not DSN-only). Numbers unquoted; string vars include quotes in the value (`sql_mode: "'STRICT_TRANS_TABLES'"`). Pin `innodb_lock_wait_timeout: "5"`.

**Server** — `maxSelectExecutionTime` caps non-transactional `Query`/`QueryRow` via `MAX_EXECUTION_TIME` (0 = off). Per-route: `service.MaxSelectExecutionTime(d)`. MySQL 3024 → HTTP 408. CLI never gets this middleware. See [database.md](database.md).

**Logging** — level, format, `logSQLQueries`.

**Propagation:**

```yaml
propagation:
  async: false              # true → dispatch propagation_requested (needs event.dispatcher=sqs)
  disableForResults: false
  logChunkCounters: true
```

If `async: true` without SQS dispatcher, scheduling logs an error and skips (no sync fallback). Startup rejects legacy `server.propagation_endpoint` / `server.disableResultsPropagation` (`ALGOREA_SERVER__*`) — a silent ignore would leave async false and run work inside API requests. Migrate to `propagation.async` + `event.dispatcher: sqs` and `propagation.disableForResults`. Commands that only `LoadConfig`/`DBConfig` (e.g. `db-migrate`) skip this check.

Ops: with `async: true`, run periodic `propagation` as a safety net (0.5s lock wait leaves a residual lost-wakeup window). Diagnostics: `performance_schema=ON`, `GRANT PROCESS ON *.*` — [propagation.md](propagation.md).

**Auth / token / domains** — login-module client, JWT keys, domain group IDs.

**CORS** — [cors.md](cors.md).
