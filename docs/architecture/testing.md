# Testing

## Kinds

| Kind | Tag / location | Role |
|------|----------------|------|
| Unit | `//go:build unit` | No DB |
| Integration | `//go:build !unit`, `*_integration_test.go` | DB + `testify`; fixtures via `testhelpers.SetupDBWithFixtureString()` |
| BDD | `*.feature` + godog | API behaviour |

## Feature file naming

| Pattern | Role |
|---------|------|
| `<operation>.feature` | Happy paths |
| `<operation>.robustness.feature` | Authz/authn failures, validation, rule violations |
| `<operation>.<aspect>.feature` | Split by aspect (`access`, `pagination`, `visibility`, …) when large |

## Commands

```bash
make test              # race + coverage → test-results/coverage.txt
make test-dev          # faster, uses cache (no coverage profile)
make test-unit
make test-bdd
make test-bdd DIRECTORY=./app/api/groups TAGS=wip
```

Godog: features next to handlers; steps in `testhelpers/steps_*.go`; lifecycle setup DB → freeze time → run → verify propagation → restore.

## Test DB

Port **3307** (`db_test`). Config: `conf/config.test.yaml` (must not point at 3306). Init once / after schema change:

```bash
docker-compose up -d db db_test
ALGOREA_ENV=test ./bin/AlgoreaBackend db-restore
ALGOREA_ENV=test ./bin/AlgoreaBackend db-migrate
ALGOREA_ENV=test ./bin/AlgoreaBackend install
```

DB connection errors → `docker-compose ps`. Known flake: a timeout test in `app/database` may hang; not a real regression.

## Coverage (required after code changes)

Touched functions and branches must hit **100%**. `make test` writes coverage; `test-dev` does not.

```bash
go test -gcflags=all=-l \
  -coverpkg=<package-under-test> \
  -coverprofile=/tmp/cov.txt -covermode=atomic \
  ./path/to/package/ -p 1 -parallel 1
go tool cover -func=/tmp/cov.txt | grep <changed-file>.go
go tool cover -html=/tmp/cov.txt -o /tmp/cov.html   # line view
```

Calling a new helper from a unit test does **not** cover the production call site. Exercise every modified caller (public entrypoint: `New`, `Reset`, HTTP handler, …). Codecov patch coverage fails on uncovered caller lines even when the helper is at 100%. Unreachable-only lines: document why; otherwise add the missing test.
