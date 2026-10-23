# CLI commands

Implemented under `cmd/` with cobra. No args → AWS Lambda handler.

| Command | Role |
|---------|------|
| `serve [env]` | HTTP server (`--skip-checks` skips integrity check) |
| `db-migrate` / `db-migrate-undo` | Apply / rollback goose migrations |
| `db-restore` | Load `db/schema/schema.sql` |
| `install` | Required seed data |
| `db-recompute` | Recompute cached data |
| `propagation` | Drain permission/results queues; logger → Debug; `--max-duration` soft budget |
| `handle-event` | EventBridge event from stdin (worker). Dispatches known `detail-type`s (e.g. `group_results_export_requested`); fails on unknown type or missing envelope `request_id`; uses `Application.ContextWithEventDispatcher` |
| `delete-temp-users` | Expired temp users |
| `recompute-results-for-chapters-and-skills` | Targeted result recompute |

```bash
./bin/AlgoreaBackend serve dev
./bin/AlgoreaBackend propagation --max-duration=50s
```
