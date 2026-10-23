# Agent instructions

You are an expert in Golang and backend development. Prefer functional, maintainable, performant Go that matches this repo's patterns.

**Architecture map:** [`ARCHITECTURE.md`](ARCHITECTURE.md) — folders, request flow, conventions, deep-dive table. Update the map for global structure; update the matching `docs/architecture/*.md` deep-dive when only that branch changes.

## Gates

- Lint must pass (`make lint` / `./bin/golangci-lint run -v --timeout 2m`). Never disable linter rules.
- Tests must pass (`make test-dev`). After code changes, touched functions/branches need **100%** coverage — recipe and DB setup in [`docs/architecture/testing.md`](docs/architecture/testing.md).

## Comments

Write **why-comments** only: non-obvious choice, hardcoded threshold, library/DB quirk, intentional omission. Skip narrating comments. On refactor, **preserve** existing explanatory comments (move with the code); drop only when factually obsolete.

## Branch docs (read when in scope)

| When | Doc |
|------|-----|
| Business rules / cross-service Algorea behavior | [domain.md](docs/agents/domain.md) (glossary, ADRs, knowledge base `docs/algorea/`) |
| Migrations / `schema.sql` / `display_settings` | [schema.md](docs/architecture/schema.md) |
| New or changed domain events | [events.md](docs/architecture/events.md) |
| Propagation / permissions_generated / results queues | [propagation.md](docs/architecture/propagation.md) |
| Config keys, async propagation, CORS | [configuration.md](docs/architecture/configuration.md), [cors.md](docs/architecture/cors.md) |

## Agent skills

### Issue tracker

Markdown under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: root `GLOSSARY.md` + `docs/adr/` (+ Algorea knowledge base). See `docs/agents/domain.md`.
