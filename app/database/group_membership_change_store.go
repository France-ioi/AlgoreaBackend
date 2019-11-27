package database

// GroupMembershipChangeStore implements database operations on `group_membership_changes`
// (which stores the history of group membership changes).
type GroupMembershipChangeStore struct {
	*DataStore
}
