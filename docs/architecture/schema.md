# Schema & migrations

## Migrations

Live in `db/migrations/` (goose). SQL or Go. Name: `YYMMDDHHMM_description.sql` (10-digit timestamp).

`db/schema/schema.sql` is the **base snapshot** for `db-restore`. `db-migrate` applies migrations on top. Do **not** hand-edit `schema.sql` for evolution — the migration file is the source of truth. Never rewrite old migrations; add new ones.

## Core tables

**Users & groups:** `users`, `groups`, `groups_groups`, `groups_ancestors`

**Items:** `items` (incl. `display_settings`), `items_strings`, `items_items`, `items_ancestors`, `item_dependencies`

**Results:** `attempts`, `results`, `answers`

**Permissions:** `permissions_granted`, `permissions_generated`

**Threads / sessions:** `threads`, `sessions`, `access_tokens`

**Propagation:** `results_propagate` (inbox; `AUTO_INCREMENT` `id` PK only), `results_propagate_internal`

**Views:** `groups_groups_active`, `groups_ancestors_active`

**InnoDB / charset:** `innodb_lock_wait_timeout=5` (also via `database.sessionParams`), `innodb_ft_min_token_size=1`, `max-allowed-packet=10485760`, ≥2GB memory; `utf8mb4` / `utf8mb4_0900_ai_ci`. Ops details also in `README.md`.

## Display settings (frontend pass-through)

`items.display_settings` (JSON, `NOT NULL`, default `{}`) is an opaque blob: stored and returned verbatim, never inspected by backend logic. Lets frontends keep UI-only settings without a schema change per key.

| Endpoint | Where |
|----------|--------|
| `GET /items/{item_id}` (`itemView`) | top-level on the item |
| `GET /items/{item_id}/children` (`itemChildrenView`) | per visible child |
| `GET /items/{item_id}/navigation` (`itemNavigationView`) | root + per child |
| `GET /current-user/group-memberships/activities` (`activitiesView`) | per root activity |
| `GET /current-user/group-memberships/skills` (`skillsView`) | per root skill |

Not exposed (yet): `itemParentsView`, `itemPrerequisitesView`, `itemDependenciesView` — re-fetch via `itemView` if needed. Rolling out is plumbing on `RawCommonItemFields`; deferred until a concrete frontend need.

Known keys (informational; backend does not enforce):

| Key | Type | Default | Notes |
|-----|------|---------|-------|
| `children_layout` | `"List" \| "Grid" \| "Hide"` | `"List"` | Children list layout |
| `prompt_to_join_group_by_code` | boolean | `false` | Show join-by-code form |

**Omit defaults:** omit a key when it equals the default (`{}` ≡ both defaults). Backend stores whatever object is sent and returns the column as stored.
