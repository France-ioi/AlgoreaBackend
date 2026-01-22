# AlgoreaBackend Architecture

**This file is mainly targeted to agents.**
**Last Updated**: 2026-01-19

## Table of Contents

1. [Project Overview](#project-overview)
2. [Technology Stack](#technology-stack)
3. [Project Structure](#project-structure)
4. [Application Architecture](#application-architecture)
5. [Database Layer](#database-layer)
6. [API Layer](#api-layer)
7. [Authentication & Authorization](#authentication--authorization)
8. [Service Layer Pattern](#service-layer-pattern)
9. [Data Propagation System](#data-propagation-system)
10. [Event Dispatch System](#event-dispatch-system)
11. [Testing Framework](#testing-framework)
12. [Configuration Management](#configuration-management)
13. [CLI Commands](#cli-commands)
14. [Database Schema Overview](#database-schema-overview)
15. [Development Workflow](#development-workflow)
16. [Key Conventions and Patterns](#key-conventions-and-patterns)

---

## Project Overview

**AlgoreaBackend** is a Go-based REST API backend for an educational platform (Algorea - France IOI). It manages:
- Users and groups (with hierarchical relationships)
- Educational items (tasks, chapters, skills)
- User progress and results
- Permissions and access control
- Attempts and submissions
- Discussion threads for help requests

The platform emphasizes:
- Complex permission propagation
- Results computation and propagation through item/group hierarchies
- Multi-domain support
- Integration with external authentication platforms

---

## Technology Stack

### Core Technologies
- **Language**: Go 1.20
- **Database**: MySQL 8.0.34
- **HTTP Router**: `go-chi/chi` v3
- **ORM**: `jinzhu/gorm` v1 (wrapped with custom extensions)
- **Testing**:
  - Standard Go testing
  - `godog` (BDD with Gherkin)
  - `testify` for assertions
- **Configuration**: `spf13/viper`
- **CLI**: `spf13/cobra`
- **Logging**: `sirupsen/logrus`
- **Migration**: `pressly/goose` v3

### Key Libraries
- `go-sql-driver/mysql` - MySQL driver
- `go-chi/render` - HTTP response rendering
- `SermoDigital/jose` - JWT handling
- `go-playground/validator` - Validation
- `luna-duclos/instrumentedsql` - SQL instrumentation for logging

### Deployment
- Docker support with `docker-compose.yml`
- AWS Lambda support via `akrylysov/algnhsa`

---

## Project Structure

```
AlgoreaBackend/
├── app/                          # Main application code
│   ├── api/                      # API endpoints grouped by resource
│   │   ├── answers/              # Answer submission endpoints
│   │   ├── auth/                 # Authentication endpoints
│   │   ├── currentuser/          # Current user endpoints
│   │   ├── groups/               # Group management endpoints
│   │   ├── items/                # Item (task/chapter) endpoints
│   │   ├── threads/              # Discussion thread endpoints
│   │   ├── users/                # User management endpoints
│   │   └── api.go                # API router setup
│   ├── auth/                     # Authentication middleware & logic
│   ├── database/                 # Database layer (stores, models, queries)
│   │   ├── configdb/             # DB config validation
│   │   ├── mysqldb/              # MySQL-specific utilities
│   │   ├── *_store.go            # Data access objects (stores)
│   │   ├── data_store.go         # DataStore wrapper
│   │   └── db.go                 # DB connection & transaction management
│   ├── domain/                   # Multi-domain configuration
│   ├── logging/                  # Structured logging utilities
│   ├── loginmodule/              # Login module integration
│   ├── payloads/                 # Request/response payload validation
│   ├── service/                  # Service layer utilities
│   │   ├── base.go               # Base service structure
│   │   ├── handler.go            # HTTP handler wrapper (error handling)
│   │   ├── errors.go             # API error definitions
│   │   ├── parameters.go         # URL/query parameter parsing
│   │   ├── query_limiter.go      # Pagination utilities
│   │   └── ...
│   ├── token/                    # Token generation & validation
│   ├── app.go                    # Application initialization
│   ├── config.go                 # Configuration loading
│   └── server.go                 # HTTP server setup
├── cmd/                          # CLI commands
│   ├── serve.go                  # Start HTTP server
│   ├── db_migrate.go             # Run migrations
│   ├── db_restore.go             # Restore DB from schema
│   ├── install.go                # Install initial data
│   ├── propagation.go            # Trigger propagation
│   └── root.go                   # Root command (Lambda handler)
├── conf/                         # Configuration files
│   ├── config.sample.yaml        # Sample config
│   ├── config.test.sample.yaml   # Sample test config
│   └── config.yml                # Active config (gitignored)
├── db/                           # Database files
│   ├── migrations/               # DB migration files
│   └── schema/                   # Base schema SQL
├── golang/                       # Generic Go utilities (Set, If, etc.)
├── testhelpers/                  # BDD test helpers
├── docker-compose.yml            # Docker setup (db + backend)
├── Dockerfile                    # Backend Docker image
├── Makefile                      # Build & test commands
└── main.go                       # Entry point
```

---

## Application Architecture

### Application Lifecycle

1. **Entry Point**: `main.go` → `cmd.Execute()`
2. **Command Routing**: Cobra handles command routing
   - Default (no args): Lambda handler
   - `serve [env]`: HTTP server
   - `db-migrate`, `db-restore`, `install`, etc.
3. **Application Initialization** (`app.New()`):
   - Load configuration from files and env vars
   - Initialize logger
   - Open DB connection
   - Set up HTTP router with middlewares
   - Mount API routes
4. **Server Start** (`app.NewServer()` → `server.Start()`):
   - Start HTTP listener (default: `:8080`)
   - Graceful shutdown support

### Request Flow

```
HTTP Request
    ↓
[Middleware Stack]
    - RealIP
    - RequestID
    - Compression (optional)
    - Structured Logger
    - Recoverer (panic handler)
    - CORS
    - Domain Middleware (sets domain config in context)
    ↓
[API Router] (/api.go)
    ↓
[Service Router] (e.g., /groups/groups.go)
    - Content-Type: application/json
    - UserMiddleware (authentication)
    ↓
[Service Handler] (e.g., /groups/get_group.go)
    - Extract parameters
    - Get user from context
    - Get DataStore from context
    - Business logic
    - Database operations
    - Return response or error
    ↓
[AppHandler Wrapper] (service/handler.go)
    - Catches panics and errors
    - Converts to APIError
    - Logs errors
    - Renders JSON response
    ↓
HTTP Response
```

### Middleware Stack

1. **LoggerMiddleware**: Adds logger to context
2. **DataStoreMiddleware**: Adds DataStore to context
3. **VersionHeaderMiddleware**: Adds X-Version header
4. **RealIP**: Extracts real IP from headers
5. **Compression**: Gzip compression (if enabled)
6. **RequestID**: Generates unique request ID
7. **StructuredLogger**: Logs requests with structured fields
8. **Recoverer**: Catches panics and returns 500
9. **CORS**: Handles CORS headers
10. **DomainMiddleware**: Sets domain-specific config in context
11. **UserMiddleware**: Authenticates user (per-service)

---

## Database Layer

### DB Connection Management

The database layer uses a custom wrapper around GORM v1 (`database.DB`) that provides:

- **Context-aware connections**: Each request gets a DB connection with request context
- **Transaction management**: Automatic retries on deadlocks/timeouts
- **Query logging**: Configurable SQL query logging
- **Connection pooling**: Via standard `database/sql`
- **Custom wrappers**: `sqlDBWrapper`, `sqlTxWrapper`, `sqlConnWrapper` for enhanced functionality

### DataStore Pattern

**`DataStore`** is the main interface for database operations:

```go
type DataStore struct {
    *DB
    tableName string
}
```

**Store Methods** (factory pattern):
```go
store.Users()           // *UserStore
store.Groups()          // *GroupStore
store.Items()           // *ItemStore
store.Results()         // *ResultStore
store.Permissions()     // *PermissionGeneratedStore
// ... and many more
```

Each store wraps a `DataStore` and provides domain-specific methods:
- Query builders
- Complex aggregations
- Business logic queries
- Trigger helpers

### Transaction Management

```go
// Basic transaction
store.InTransaction(func(s *DataStore) error {
    // Operations within transaction
    return nil
})

// Ensures transaction (uses existing if already in one)
store.EnsureTransaction(func(s *DataStore) error {
    // ...
})

// With foreign key checks disabled
store.WithForeignKeyChecksDisabled(func(s *DataStore) error {
    // ...
})

// With named lock
store.WithNamedLock("lock_name", timeout, func(s *DataStore) error {
    // ...
})
```

**Key Features**:
- Automatic retry on deadlocks/timeouts (up to 30 retries)
- Nested transaction support via EnsureTransaction
- Context cancellation support
- Row-level locking: `WithExclusiveWriteLock()`, `WithSharedWriteLock()`

### Common Table Expressions (CTEs)

Custom CTE support via `With()`:

```go
query := store.
    With("user_ancestors", ancestorsQuery.SubQuery()).
    Joins("JOIN user_ancestors ON ...")
```

### Key Database Utilities

- **`InsertMap/InsertMaps`**: Insert from map[string]interface{}
- **`InsertOrUpdateMaps`**: Upsert with ON DUPLICATE KEY UPDATE
- **`ScanIntoSlices`**: Scan multiple columns into slices
- **`ScanIntoSliceOfMaps`**: Scan rows into []map[string]interface{}
- **`PluckFirst`**: Get first value of a column
- **`HasRows`**: Check if query returns any rows
- **`RetryOnDuplicateKeyError`**: Retry on duplicate key (for ID generation)

---

## API Layer

### Service Organization

Each resource has its own package under `app/api/`:
- `answers/` - Answer submissions
- `auth/` - Authentication (login, logout, refresh)
- `currentuser/` - Current user info & settings
- `groups/` - Group CRUD, permissions, members, managers
- `items/` - Item CRUD, navigation, dependencies, breadcrumbs
- `threads/` - Discussion threads for help requests
- `users/` - User search and info

### Service Structure

Each service package contains:
1. **`<resource>.go`**: Service struct + route definitions
2. **`<operation>.go`**: Handler implementations (e.g., `get_group.go`, `update_group.go`)
3. **`*.feature`**: BDD test scenarios (Gherkin)
4. **`bdd_test.go`**: BDD test entry point

Example (`groups/groups.go`):

```go
type Service struct {
    *service.Base
}

func (srv *Service) SetRoutes(router chi.Router) {
    router.Use(render.SetContentType(render.ContentTypeJSON))
    router.Use(auth.UserMiddleware(srv.Base))
    router.Get("/groups/{group_id}", service.AppHandler(srv.getGroup).ServeHTTP)
    router.Put("/groups/{group_id}", service.AppHandler(srv.updateGroup).ServeHTTP)
    // ... more routes
}
```

### Handler Pattern

```go
func (srv *Service) getGroup(w http.ResponseWriter, r *http.Request) error {
    // 1. Parse parameters
    groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
    if err != nil {
        return service.ErrInvalidRequest(err)
    }

    // 2. Get user and store from context
    user := srv.GetUser(r)
    store := srv.GetStore(r)

    // 3. Check permissions
    var group GroupResponse
    err = store.Groups().PickVisibleGroups(store.Groups().ByID(groupID), user).
        Scan(&group).Error()
    if gorm.IsRecordNotFoundError(err) {
        return service.ErrAPIInsufficientAccessRights
    }
    service.MustNotBeError(err)

    // 4. Return response
    render.Respond(w, r, group)
    return nil
}
```

### Error Handling

All errors are wrapped in `APIError`:

```go
// Defined in service/errors.go
ErrInvalidRequest(err)              // 400 Bad Request
ErrUnauthorized(reason)              // 401 Unauthorized
ErrAPIInsufficientAccessRights       // 403 Forbidden
ErrForbidden(reason)                 // 403 Forbidden
ErrNotFound(resource)                // 404 Not Found
ErrRequestTimeout()                  // 408 Request Timeout
ErrUnexpected(err)                   // 500 Internal Server Error
```

Errors are automatically converted to JSON responses by `AppHandler`.

### Response Rendering

Uses `go-chi/render` with custom responder (`service.AppResponder`):
- Automatically sets `success: true` on successful responses
- Handles pagination metadata
- Renders errors with `success: false, message: "..."`

---

## Authentication & Authorization

### Authentication Flow

1. **Access Token Validation**:
   - Token can come from `Authorization: Bearer <token>` header or `access_token` cookie
   - `auth.UserMiddleware` validates token against `sessions` table
   - Loads user from `users` table
   - Adds user to request context

2. **Session Management**:
   - Sessions stored in `sessions` table with expiration
   - Access tokens linked to sessions
   - Refresh tokens supported for token renewal
   - Temporary users have special handling

3. **User Context**:
   ```go
   user := srv.GetUser(r)          // *database.User
   sessionID := srv.GetSessionID(r) // int64
   ```

### Authorization Patterns

#### Group Permissions
- Hierarchical group structure (groups_groups, groups_ancestors)
- Manager permissions with granular capabilities
- Visibility checks via `PickVisibleGroups()`

#### Item Permissions
- View permissions: `none`, `info`, `content`, `content_with_descendants`, `solution`
- Watch permissions: `none`, `result`, `answer`, `answer_with_grant`
- Edit/grant permissions
- Permissions computed and cached in `permissions_generated` table

#### Permission Checking
```go
// Check if user can view an item
store.Items().PickVisibleItems(store.Items().ByID(itemID), user)

// Check if user can manage a group
checkThatUserCanManageTheGroup(store, groupID, user, canManage)
```

### Domain-Based Configuration

Multi-domain support for different user bases:
```yaml
domains:
  - domains: ["algorea.org", "france-ioi.org"]
    all_users_group: 1
    non_temp_users_group: 2
    temp_users_group: 3
```

Retrieved via:
```go
domainConfig := domain.ConfigFromContext(r.Context())
```

---

## Service Layer Pattern

### Base Service

All services embed `service.Base`:

```go
type Base struct {
    store        *database.DataStore  // Global store
    ServerConfig *viper.Viper
    AuthConfig   *viper.Viper
    DomainConfig []domain.ConfigItem
    TokenConfig  *token.Config
}
```

**Methods**:
- `GetUser(r)`: Get authenticated user
- `GetSessionID(r)`: Get session ID
- `GetStore(r)`: Get request-scoped DataStore

### Utility Functions

**Parameter Parsing** (`service/parameters.go`):
```go
ResolveURLQueryPathInt64Field(r, "group_id")
ResolveURLQueryGetInt64Field(r, "parent_group_id")
ResolveURLQueryGetBoolField(r, "descendants", false)
```

**Query Limiting** (`service/query_limiter.go`):
```go
limiter := service.NewQueryLimiter()
query = limiter.Apply(r, query)  // Applies ?limit and ?offset
```

**Sorting** (`service/sorting.go`):
```go
sorter := service.NewSortBuilder()
query = sorter.Apply(r, query, sortingFields)
```

**Propagation** (`service/propagation.go`):
```go
service.SchedulePropagation(store, r, "groups")  // Async propagation
service.MustPropagateNow(r, store)               // Sync propagation
```

---

## Data Propagation System

One of the most complex parts of the system. It ensures that:
- Group hierarchies are correctly computed
- Permissions cascade through groups and items
- Results aggregate up through item hierarchies

### Propagation Types

1. **Permissions Propagation**:
   - Triggered by: group relationship changes, permission grants
   - Computes `permissions_generated` from `permissions_granted` + group/item hierarchies
   - Uses `PermissionGrantedStore.computeAllAccess()`

2. **Results Propagation**:
   - Triggered by: new results, result updates, item relationship changes
   - Aggregates child results up to parent items
   - Unlocks items based on dependencies
   - Uses `ResultStore.propagate()`

### Propagation Scheduling

**Async Mode** (default):
```go
store.SchedulePermissionsPropagation()
store.ScheduleResultsPropagation()
```

- Propagation runs **after transaction commit**
- Each propagation type runs once per transaction
- Uses separate transactions for each propagation step

**Sync Mode** (for specific operations):
```go
store.SetPropagationsModeToSync()
```

- Propagation runs **before transaction commit**
- Used when results need to be immediately visible

### Results Propagation Algorithm

1. **Mark for propagation**: Rows in `results_propagate` table
2. **Move to internal**: `results_propagate` → `results_propagate_internal`
3. **Process in chunks**:
   - Mark chunk as 'propagating'
   - Mark parents as 'to_be_recomputed'
   - Unlock items based on dependencies
   - Recompute aggregates (score, validation, etc.)
   - Mark changed results as 'to_be_propagated'
4. **Repeat** until no more results to propagate

**Concurrency Control**:
- Named lock: `results_propagation` (10s timeout)
- Prevents parallel propagation
- Ensures consistency

**Recomputed Fields**:
- `latest_activity_at`
- `tasks_tried`, `tasks_with_help`
- `validated_at`
- `score_computed`

---

## Event Dispatch System

The event dispatch system sends domain events to external systems (e.g., AWS SQS) for further processing.

### Architecture

```
HTTP Handler → Transaction Commits → Dispatch Event
                                          ↓
                                    Dispatcher (SQS or NoOp)
                                          ↓
                                    AWS SQS → EventBridge
```

### Event Structure

```go
type Event struct {
    Version   string                 `json:"version"`    // Schema version (e.g., "1.0")
    Type      string                 `json:"type"`       // Event type (e.g., "submission_created")
    SourceApp string                 `json:"source_app"` // Always "algoreabackend"
    Instance  string                 `json:"instance"`   // Optional (e.g., "prod", "staging")
    Time      time.Time              `json:"time"`       // When the event occurred
    RequestID string                 `json:"request_id"` // For correlation with logs
    Payload   map[string]interface{} `json:"payload"`    // Event-specific data
}
```

### Event Types

- `submission_created`: User submitted an answer
- `grade_saved`: A grade was saved for an answer
- `item_unlocked`: Item was unlocked for a user
- `thread_status_changed`: Help thread status changed

### Configuration

```yaml
event:
  dispatcher: "sqs"        # "sqs" or empty for no-op
  instance: "prod"         # Optional instance identifier
  sqs:
    queueURL: "https://sqs.eu-west-1.amazonaws.com/..."
    region: "eu-west-1"
```

### Usage in Handlers

```go
// After transaction commits
event.Dispatch(httpRequest.Context(), event.TypeSubmissionCreated, map[string]interface{}{
    "author_id":      userID,
    "participant_id": participantID,
    "item_id":        itemID,
    "attempt_id":     attemptID,
    "answer_id":      answerID,
})
```

### Key Design Decisions

- **Timing**: Events are dispatched synchronously after transaction commits (required for Lambda)
- **Error Handling**: Dispatch errors are logged but don't fail the request (best-effort)
- **Timeout**: SQS calls have a 1-second timeout
- **Testing**: Mock dispatcher is injected via context for BDD tests

### Versioning

Event schema versions follow semver-like rules:
- **Minor bump** (1.0 → 1.1): Non-breaking changes (adding optional fields)
- **Major bump** (1.x → 2.0): Breaking changes (removing fields, changing semantics)

---

## Testing Framework

### Test Types

1. **Unit Tests**: Standard Go tests tagged with `//go:build unit`
2. **Integration Tests**: Tests requiring DB, tagged with `//go:build !unit`
3. **BDD Tests**: Gherkin feature files tested with godog

### Gherkin Feature File Categories

Feature files follow naming conventions to organize tests by purpose:

#### 1. Regular Feature Files (`<operation>.feature`)

Test the **main functionality** and expected behaviors (happy paths):
- Successful API responses
- Valid use cases with correct permissions
- Expected data transformations

**Example**: `get_thread.feature`
```gherkin
Scenario: Should return all fields when the thread exists
  Given I am the user with id "1"
  When I send a GET request to "/items/21/participant/1/thread"
  Then the response code should be 200
  And the response body should be, in JSON:
    """
    { "participant_id": "1", "item_id": "21", "status": "waiting_for_trainer" }
    """
```

#### 2. Robustness Feature Files (`<operation>.robustness.feature`)

Test **error handling, edge cases, and failure scenarios**:
- Authentication failures (missing/invalid/expired tokens)
- Authorization failures (insufficient permissions)
- Input validation errors (wrong types, missing parameters)
- Business rule violations

**Example**: `get_thread.robustness.feature`
```gherkin
Scenario: Should be logged
  When I send a GET request to "/items/10/participant/1/thread"
  Then the response code should be 401
  And the response error message should contain "No access token provided"

Scenario: The item_id parameter should be an int64
  Given I am the user with id "1"
  When I send a GET request to "/items/aaa/participant/1/thread"
  Then the response code should be 400
  And the response error message should contain "Wrong value for item_id (should be int64)"
```

#### 3. Split Feature Files (`<operation>.<aspect>.feature`)

When an endpoint has many test scenarios, split by **aspect being tested**:
- `<operation>.access.feature` - Access control and permission scenarios
- `<operation>.pagination.feature` - Pagination behavior
- `<operation>.visibility.feature` - Visibility rules for returned data

**Example**: `get_permission_explanation` is split into:
- `get_permission_explanation.feature` - Main functionality
- `get_permission_explanation.robustness.feature` - Error cases
- `get_permission_explanation.access.feature` - Access control scenarios
- `get_permission_explanation.pagination.feature` - Pagination scenarios
- `get_permission_explanation.visibility.feature` - Visibility rules

### Go Integration Tests (`*_integration_test.go`)

Standard Go tests (not Gherkin) that require database access for testing internal functions:
- Use `//go:build !unit` tag
- Use `testify` for assertions
- Test complex internal logic that isn't directly exposed via API
- Use YAML fixtures loaded via `testhelpers.SetupDBWithFixtureString()`

**Example**: `create_invitations_integration_test.go`
```go
func Test_filterOtherTeamsMembersOut(t *testing.T) {
    tests := []struct {
        name           string
        fixture        string
        groupsToInvite []int64
        want           []int64
    }{
        {
            name: "parent group is not a team",
            fixture: `
                groups:
                    - {id: 1, type: Class}
                    - {id: 10, type: User}
                groups_groups: [{parent_group_id: 2, child_group_id: 10}]`,
            groupsToInvite: []int64{10},
            want:           []int64{10},
        },
    }
    // ... test implementation
}
```

### Running Tests

```bash
make test              # All tests with race detection & coverage
make test-dev          # All tests, faster (uses cache)
make test-unit         # Only unit tests
make test-bdd          # Only BDD tests
make test-bdd DIRECTORY=./app/api/groups TAGS=wip  # Filtered BDD
```

### BDD Testing with Godog

**Structure**:
- Feature files: `*.feature` (Gherkin syntax)
- Step definitions: `testhelpers/steps_*.go`
- Test context: `testhelpers/test_context.go`

**Example Feature** (`items/get_breadcrumbs.feature`):
```gherkin
Feature: Get item breadcrumbs

Background:
  Given the database has the following table "groups":
    | id | name    | type  |
    | 11 | jdoe    | User  |
  And the database has the following user:
    | group_id | login | default_language |
    | 11       | jdoe  | fr               |
  And I am the user with id "11"

Scenario: Full access on all the breadcrumbs
  Given I can view content of the item @item1
  When I send a GET request to "/items/21/breadcrumbs"
  Then the response code should be 200
  And the response body should be, in JSON:
    """
    [
      {"id": "21", "title": "Graphe: Methodes"}
    ]
    """
```

**Step Definitions** (`testhelpers/`):
- `steps_db.go`: Database setup steps
- `steps_request.go`: HTTP request steps
- `steps_response.go`: Response assertion steps
- `app_language_*.go`: Domain-specific step helpers

**Test Lifecycle**:
1. **Before**: Setup test DB, freeze time, create app
2. **Steps**: Execute Gherkin steps
3. **After**: Verify propagation, restore DB, cleanup

### Test Database

- Uses separate test DB on port `3307` (via `db_test` service)
- Config: `conf/config.test.yaml`
- Initialized with: `ALGOREA_ENV=test ./bin/AlgoreaBackend db-restore && db-migrate && install`

---

## Configuration Management

### Configuration Loading

**Priority** (highest to lowest):
1. Explicit `Set()` calls
2. Command-line flags
3. Environment variables
4. Config file
5. Defaults

**Files**:
- `conf/config.yml` (base config, not in test env)
- `conf/config.{env}.yml` (environment-specific)

**Environment Variables**:
- Prefix: `ALGOREA_`
- Nested keys: double underscore `__`
- Example: `ALGOREA_DATABASE__ADDR=localhost:3306`

### Configuration Sections

**Database** (`database`):
```yaml
database:
  addr: localhost:3306
  user: algorea
  passwd: a_db_password
  dbname: algorea_db
```

**Server** (`server`):
```yaml
server:
  rootPath: /
  compress: true
  disableResultsPropagation: false
```

**Logging** (`logging`):
```yaml
logging:
  level: debug
  output: stdout
  format: json
  logSQLQueries: true
```

**Auth** (`auth`):
- Client ID/secret for login module
- Token lifetimes

**Token** (`token`):
- Public/private keys for JWT
- Encryption keys

**Domains** (`domains`):
- Domain-specific group IDs

### Accessing Configuration

```go
config := app.LoadConfig()
dbConfig, _ := app.DBConfig(config)
serverConfig := app.ServerConfig(config)
domainConfig, _ := app.DomainsConfig(config)
```

---

## CLI Commands

### Available Commands

**Default** (no args): AWS Lambda handler

**`serve [env]`**: Start HTTP server
```bash
./bin/AlgoreaBackend serve dev
./bin/AlgoreaBackend serve --skip-checks  # Skip integrity check
```

**`db-migrate`**: Run pending migrations
```bash
./bin/AlgoreaBackend db-migrate
```

**`db-migrate-undo`**: Rollback last migration

**`db-restore`**: Restore DB from schema file
```bash
./bin/AlgoreaBackend db-restore
```

**`install`**: Install required initial data
```bash
./bin/AlgoreaBackend install
```

**`db-recompute`**: Recompute cached data
```bash
./bin/AlgoreaBackend db-recompute
```

**`propagation`**: Trigger propagation manually

**`delete-temp-users`**: Delete expired temporary users

**`recompute-results-for-chapters-and-skills`**: Recompute specific results

### Command Implementation

Each command is defined in `cmd/<command>.go`:
- Uses `cobra` for command structure
- Initializes app/DB as needed
- Performs operation
- Returns error on failure

---

## Database Schema Overview

### Core Tables

**Users & Groups**:
- `users`: User profiles, linked to auth platform
- `groups`: Groups of any type (User, Team, Class, etc.)
- `groups_groups`: Parent-child group relationships
- `groups_ancestors`: Computed transitive closure of group relationships

**Items (Educational Content)**:
- `items`: Tasks, chapters, skills, courses
- `items_strings`: Localized item content (title, description, etc.)
- `items_items`: Parent-child item relationships
- `items_ancestors`: Computed transitive closure of item relationships
- `item_dependencies`: Unlock conditions between items

**Results & Attempts**:
- `attempts`: User attempts on items
- `results`: User progress on items (per attempt)
- `answers`: User submissions/answers

**Permissions**:
- `permissions_granted`: Explicitly granted permissions
- `permissions_generated`: Computed effective permissions (cached)

**Threads**:
- `threads`: Help request discussions

**Sessions**:
- `sessions`: User sessions
- `access_tokens`: API access tokens

**Propagation**:
- `results_propagate`: Results marked for propagation
- `results_propagate_internal`: Internal propagation queue

### Key Views

- `groups_groups_active`: Active (non-expired) group relations
- `groups_ancestors_active`: Active group ancestor closure

### Database Constraints

**InnoDB Settings**:
- `innodb_lock_wait_timeout=5`
- `innodb_ft_min_token_size=1`
- `max-allowed-packet=10485760`
- Memory: >= 2GB

**Character Set**: `utf8mb4` with `utf8mb4_0900_ai_ci` collation

---

## Development Workflow

### Setup

1. **Clone repo**
2. **Start databases**: `docker-compose up -d db db_test`
3. **Copy config**: `cp conf/config.sample.yaml conf/config.yml`
4. **Copy test config**: `cp conf/config.test.sample.yaml conf/config.test.yaml`
5. **Build**: `make`
6. **Restore DB**: `./bin/AlgoreaBackend db-restore`
7. **Migrate**: `./bin/AlgoreaBackend db-migrate`
8. **Install data**: `./bin/AlgoreaBackend install`

### Build & Run

```bash
make                    # Build binary
./bin/AlgoreaBackend serve  # Start server
```

### Linting

```bash
make lint
./bin/golangci-lint run -v --timeout 2m
```

**Rules**: Defined in `.golangci.yml` (not shown here, but exists)
- Disabling linter rules is **not allowed**

### Code Formatting

Uses `gofumpt` (stricter than `gofmt`)

### Git Hooks

Copy `githooks/pre-commit` to `.git/hooks/pre-commit` for automatic checks

### Profiling

```bash
# 1. Start server in dev mode
./bin/AlgoreaBackend serve dev

# 2. Generate load
ab -k -c 1 -n 10000 -H "Authorization: Bearer 1" "http://127.0.0.1:8080/groups/5/team-descendants"

# 3. Get profile
go tool pprof http://127.0.0.1:8080/debug/pprof/profile?seconds=10
```

---

## Key Conventions and Patterns

### Code Organization

1. **One handler per file**: `get_group.go`, `update_group.go`
2. **Service struct embeds `service.Base`**
3. **Routes defined in `SetRoutes()`**
4. **Stores accessed via `DataStore` factory methods**
5. **Errors returned, not panicked** (except in test helpers)

### Naming Conventions

- **Stores**: `<Table>Store` (e.g., `UserStore`, `GroupStore`)
- **Handlers**: `<verb><Resource>` (e.g., `getGroup`, `updateGroup`)
- **Files**: `<operation>.go` (e.g., `get_group.go`)
- **Tests**: `<file>_test.go` or `*.feature`

### Error Handling

- **API errors**: Return `service.ErrXxx()`
- **Unexpected errors**: `service.MustNotBeError(err)` (panics on error, caught by `AppHandler`)
- **DB errors**: Check with `gorm.IsRecordNotFoundError(err)`

### Context Usage

- **Logger**: `logging.EntryFromContext(ctx)`
- **User**: `auth.UserFromContext(ctx)`
- **SessionID**: `auth.SessionIDFromContext(ctx)`
- **Domain Config**: `domain.ConfigFromContext(ctx)`

### Transactions

- Use `InTransaction()` for all writes
- Use `EnsureTransaction()` when transaction may already exist
- Schedule propagations inside transactions
- Retry logic is automatic

### Testing

- Tag unit tests: `//go:build unit`
- Tag integration tests: `//go:build !unit`
- Use BDD for API tests
- Use table-driven tests for utilities

### SQL Queries

- Prefer query builders over raw SQL
- Use CTEs for complex queries (`With()`)
- Use `SubQuery()` for subqueries
- Quote identifiers: `database.QuoteName("table_name")`
- Escape LIKE strings: `database.EscapeLikeString(value, '\\')`

### Migrations

- Named: `YYMMDDHHMMSS_description.sql` or `.go`
- Up and down in same file (for SQL)
- Use `goose` for migration management
- Never modify old migrations (create new ones)

---

## Additional Notes

### Performance Considerations

- **Propagation**: Can be slow for large changes; uses chunking
- **Query optimization**: Add indexes carefully; monitor slow query log
- **Transaction retries**: Automatic retry on deadlocks (up to 30 times)
- **Connection pooling**: Managed by `database/sql`

### Security

- **SQL injection**: Prevented by parameterized queries
- **Authorization**: Checked in every handler via `PickVisibleGroups`, `PickVisibleItems`
- **CORS**: Configured in `app/cors.go`
- **Token validation**: In `auth.UserMiddleware`

### Observability

- **Structured logging**: JSON or text format
- **Request ID**: Added to all logs
- **SQL query logging**: Optional (set `logSQLQueries: true`)
- **Profiling**: Available in dev mode at `/debug/pprof`

### Deployment

- **Docker**: Use `docker-compose up` for local dev
- **AWS Lambda**: Supported via `algnhsa` adapter
- **Health check**: `GET /status` endpoint

---

## Future Agent Notes

When working with this codebase:

1. **Always check permissions** before allowing operations
2. **Use transactions** for write operations
3. **Schedule propagations** when modifying groups/items/permissions/results
4. **Follow the service pattern** when adding new endpoints
5. **Write BDD tests** for new API endpoints
6. **Never disable linter rules**
7. **Use `MustNotBeError()` judiciously** (only for truly unexpected errors)
8. **Test with both `make test` and `make test-dev`** before pushing
9. **Update this file** when making architectural changes
