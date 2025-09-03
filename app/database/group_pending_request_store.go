package database

// GroupPendingRequestStore implements database operations on `group_pending_requests`
// (which stores requests that require an action from a user).
type GroupPendingRequestStore struct {
	*DataStore
}
