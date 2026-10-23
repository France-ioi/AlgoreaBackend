# Database layer

## Connection & wrappers

Custom wrapper around GORM v1 (`database.DB`):

- Context-aware connections per request
- Transaction retries on deadlocks / lock-wait timeouts
- Query logging; pooling via `database/sql`
- Wrappers: `sqlDBWrapper`, `sqlTxWrapper`, `sqlConnWrapper`

MySQL session variables from `database.sessionParams` are re-applied after every `COM_RESET_CONNECTION` (DSN params alone do not survive pool reuse). On re-apply failure the pool discards the connection (`driver.ErrBadConn`).

Optional API `maxSelectExecutionTime` (config / per-route middleware): injects `MAX_EXECUTION_TIME` on non-transactional `Query`/`QueryRow` only; MySQL error 3024 is not retried and surfaces as HTTP 408. See [configuration.md](configuration.md).

## DataStore

```go
type DataStore struct {
    *DB
    tableName string
}
```

Factories: `store.Users()`, `Groups()`, `Items()`, `Results()`, `Permissions()`, …. Each store adds domain query builders, aggregations, and trigger helpers.

## Transactions

```go
store.InTransaction(func(s *DataStore) error { /* … */ return nil })
store.EnsureTransaction(func(s *DataStore) error { /* nested-safe */ return nil })
store.WithForeignKeyChecksDisabled(func(s *DataStore) error { /* … */ return nil })
store.WithNamedLock("lock_name", timeout, func(s *DataStore) error { /* … */ return nil })
```

**Retries:** up to 30 retries or a 30s wall-clock *retry* budget from the first retryable failure (whichever first). Worst case ≈ `retriesTimeBudget + innodb_lock_wait_timeout` (~35s at production’s 5s lock wait). Budget is per transaction / per autocommit statement, not per HTTP request. Lock-wait timeouts log at `warning`; deadlocks at `info`.

**Locks:** `WithExclusiveWriteLock()`, `WithSharedWriteLock()`. Named locks (`GET_LOCK`/`RELEASE_LOCK`) are namespaced by MySQL schema (needs DSN string so the schema is known; opening from `*sql.DB`/`*sql.Tx` leaves namespacing off).

**Isolation exception:** the `results_propagate` inbox drain uses `READ COMMITTED` (via `EnsureTransaction` + `TxOptions`) so `INSERT … SELECT` does not take shared next-key locks on the append-only inbox. Do not change isolation inside an already-open transaction. See [propagation.md](propagation.md).

## CTEs & helpers

```go
query := store.
    With("user_ancestors", ancestorsQuery.SubQuery()).
    Joins("JOIN user_ancestors ON ...")
```

| Helper | Role |
|--------|------|
| `InsertMap` / `InsertMaps` | Insert from maps |
| `InsertOrUpdateMaps` | Upsert (`ON DUPLICATE KEY UPDATE`) |
| `ScanIntoSlices` / `ScanIntoSliceOfMaps` | Multi-column / row maps |
| `PluckFirst` | First column value |
| `HasRows` | Existence check |
| `RetryOnDuplicateKeyError` | ID-generation retries |
