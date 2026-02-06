package event

// Event type constants.
const (
	// TypeSubmissionCreated is dispatched when a user submits an answer.
	TypeSubmissionCreated = "submission_created"

	// TypeGradeSaved is dispatched when a grade is saved for an answer.
	TypeGradeSaved = "grade_saved"

	// TypeItemUnlocked is dispatched when an item is unlocked for a user.
	TypeItemUnlocked = "item_unlocked"

	// TypeThreadStatusChanged is dispatched when a thread's status changes.
	TypeThreadStatusChanged = "thread_status_changed"

	// TypeUserAuthenticated is dispatched when a user authenticates via the login module.
	TypeUserAuthenticated = "user_authenticated"
)
