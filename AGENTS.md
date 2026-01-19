
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
