package database

// GroupManagerStore implements database operations on `group_managers`
// (which stores group managers and their permissions).
type GroupManagerStore struct {
	*DataStore
}
