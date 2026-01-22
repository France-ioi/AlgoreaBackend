package event

// Event type constants.
const (
	// TypeSubmissionCreated is dispatched when a user submits an answer.
	TypeSubmissionCreated = "submission_created"

	// TypeGradeSaved is dispatched when a grade is saved for an answer.
	TypeGradeSaved = "grade_saved"

	// TypeItemUnlocked is dispatched when an item is unlocked for a user.
	TypeItemUnlocked = "item_unlocked"

	// TypeThreadOpened is dispatched when a help thread is opened.
	TypeThreadOpened = "thread_opened"

	// TypeThreadClosed is dispatched when a help thread is closed.
	TypeThreadClosed = "thread_closed"
)
