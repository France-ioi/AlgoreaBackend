
# Agent instructions

You are an expert in Golang and backend development. You write functional, maintainable, performant, and accessible code following Golang best practices.

The architecture of the project is documented in `ARCHITECTURE.md`. Whenever changes are made to the architecture, this file must be updated accordingly.

## Linting and testing

All code must pass the linting rules and tests.

### Linting

You can run the linter with:
```
./bin/golangci-lint run -v --timeout 2m
```

Disabling existing linting rules is not allowed.

### Testing

#### Prerequisites

1. Start the database services (both `db` and `db_test`):
```
docker-compose up -d db db_test
```

2. Ensure `conf/config.test.yaml` exists and points to the test database on port `3307` (not `3306`):
```yaml
database:
  addr: localhost:3307  # Must use port 3307 for db_test
```

3. Initialize the test database (only needed once, or after database schema changes):
```
ALGOREA_ENV=test ./bin/AlgoreaBackend db-restore
ALGOREA_ENV=test ./bin/AlgoreaBackend db-migrate
ALGOREA_ENV=test ./bin/AlgoreaBackend install
```

#### Running Tests

To run all tests:
```
make test-dev
```

Note: There may be a flaky timeout test in `app/database` that occasionally hangs. This is a known issue and doesn't indicate a real problem.

If you get DB connection errors, ensure docker-compose services are running with `docker-compose ps`.

## Event System

The project has an event dispatch system that sends domain events to external systems (e.g., AWS SQS).

### Event Versioning Rules

Events include a `version` field that follows semver-like rules. When modifying event schemas:

- **Minor version bump** (1.0 → 1.1): Non-breaking changes only
  - Adding new optional fields to the payload
  - Adding new event types
  
- **Major version bump** (1.x → 2.0): Breaking changes
  - Removing fields from the payload
  - Changing field types or semantics
  - Renaming fields

The version constant is defined in `app/event/event.go` as `EventVersion`. Update it when making schema changes.

### Adding New Events

1. Add a new event type constant in `app/event/types.go`
2. Call `event.Dispatch()` after the relevant transaction commits
3. Add BDD tests to verify the event is dispatched with the correct payload
4. Update `ARCHITECTURE.md` with the new event type
