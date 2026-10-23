# Data propagation

Keeps group hierarchies, permission cascades, and result aggregates consistent.

## Types

1. **Permissions** — group relation / grant changes → `permissions_generated` via `PermissionGrantedStore.computeAllAccess()`
2. **Results** — new/updated results, item relation changes → aggregate up the item tree, unlock dependencies via `ResultStore.propagate()`

## Scheduling

**Async** (default): `SchedulePermissionsPropagation` / `ScheduleResultsPropagation` — runs after commit, once per type per transaction, each in its own transaction. When `propagation.async` is true, API scheduling dispatches `propagation_requested` (needs `event.dispatcher=sqs`); otherwise it can run sync after commit. See [configuration.md](configuration.md) and [events.md](events.md).

**Sync:** `SetPropagationsModeToSync()` — before commit, when callers need immediately visible results.

```go
service.SchedulePropagation(store, srv.IsPropagationAsync(), []string{"results"})
```

## Results algorithm

1. Mark rows in `results_propagate` (append-only inbox: `AUTO_INCREMENT` `id` PK only — no uniqueness on `(participant_id, attempt_id, item_id)`, no FK, no `state` index — so API writers do not contend with the drain)
2. Drain `results_propagate` → `results_propagate_internal` by high-water-mark chunks under `READ COMMITTED` (each id-range chunk its own short transaction)
3. Per chunk: mark `propagating` → parents `to_be_recomputed` → unlock deps → recompute aggregates → mark changed as `to_be_propagated`
4. Repeat until empty

**Soft budget:** `propagation --max-duration` stops between committed chunks (permissions + results, including `results_recompute_for_items`). Leftover rows stay queued. A bare `propagation` run drains itself (always schedules both). Quiet systems with no later trigger can leave work stale — trade-off vs Lambda mid-transaction kills.

**Post-unlock:** after unlocks, `computeAllAccess()` may run with `postUnlockPermissionsBudget` (30s), inside the CLI shutdown margin (90s). Residual permissions stay in `permissions_propagate`.

**Recomputed fields:** `latest_activity_at`, `tasks_tried`, `tasks_with_help`, `validated_at`, `score_computed`.

## Concurrency

- Named lock `results_propagation` (10s), schema-namespaced
- CLI also holds `propagation_command` (wait ≤ 0.5s, clamped to remaining soft budget)

## Observability

- Per-chunk duration: Debug, or Warn if ≥ 5s (`resultsPropagationSlowChunkThreshold`)
- When `propagation.logChunkCounters` is true: non-logging path snapshots status counters onto that line. After-snapshot skipped when chunk is fast and Debug is off. `Handler_read_key` is session-scoped; InnoDB read counters are server-wide — treat “low rows + high duration” as a hint; prefer `INNODB_TRX` dump. Needs `performance_schema=ON` (ops). Degrades permanently on empty/NULL/permanent errors; transient errors use a cooldown
- Slow chunks dump oldest long-running `information_schema.INNODB_TRX` rows (started > 30s; `LIMIT 11`) via non-logging path — needs `GRANT PROCESS ON *.*`. Access denied disables dumps for the process after one Warn
- Lock-wait retries: Warn; deadlocks: Info
- CLI `propagation` raises its own logger to Debug; HTTP API level unchanged

## Performance notes

Inbox drain walks by `id` under `READ COMMITTED` so API appends do not fight whole-table shared locks; deploy the `id` PK migration only when the inbox is empty. Prefer indexes carefully; watch the slow query log. Transaction retry budget: [database.md](database.md).
