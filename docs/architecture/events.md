# Event dispatch

Sends domain events to external systems (e.g. AWS SQS → EventBridge → worker Lambda).

```
Handler → txn commits → Dispatch → Dispatcher (SQS | NoOp) → SQS → EventBridge
                                                              → (propagation_requested) worker
```

## Envelope

```go
type Event struct {
    Version   string                 `json:"version"`
    Type      string                 `json:"type"`
    SourceApp string                 `json:"source_app"` // "algoreabackend"
    Instance  string                 `json:"instance"`
    Time      time.Time              `json:"time"`
    RequestID string                 `json:"request_id"`
    Payload   map[string]interface{} `json:"payload"`
}
```

## Types

| Type | When |
|------|------|
| `submission_created` | Answer submitted |
| `grade_saved` | Grade saved; payload `score_improved` if computed score rose |
| `item_unlocked` | Item unlocked |
| `thread_status_changed` | Help thread status changed |
| `user_authenticated` | Login-module auth (code login) |
| `propagation_requested` | Async perm/results work; payload `types` is observational — worker always runs both |
| `group_results_export_completed` | Async ZIP export finished; payload `export_id`, `status`, `token`, metadata, optional `error` (`too_many_entries`, `token_invalid`, `upload_failed`, `internal`). Worker uses `Application.ContextWithEventDispatcher` |

## Config

```yaml
event:
  dispatcher: "sqs"   # or empty → no-op
  instance: "prod"
  sqs:
    queueURL: "https://sqs.…"
    region: "eu-west-1"
```

## Usage

```go
event.Dispatch(httpRequest.Context(), event.TypeSubmissionCreated, map[string]interface{}{
    "author_id": userID, "participant_id": participantID,
    "item_id": itemID, "attempt_id": attemptID, "answer_id": answerID,
})
```

**Design:** dispatch after commit (Lambda-safe); errors logged, request still succeeds; `request_id` copied from context (`middleware.GetReqID`), never invented — HTTP from chi RequestID; `handle-event` propagates inbound envelope `request_id` (fail-fast if missing for `group_results_export_requested`); SQS timeout 1s; BDD injects mock via context.

## Versioning

`EventVersion` in `app/event/event.go`. Semver-like:

- **Minor** (1.0 → 1.1): optional fields, new event types
- **Major** (1.x → 2.0): remove/rename fields or change semantics

## Adding an event

1. Constant in `app/event/types.go`
2. `event.Dispatch()` after the relevant transaction commits
3. BDD asserting payload
4. Update the **Types** table in this file
