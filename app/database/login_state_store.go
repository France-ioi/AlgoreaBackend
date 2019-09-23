package database

// LoginStateStore implements database operations on login_states
type LoginStateStore struct {
	*DataStore
}

// DeleteExpired deletes all expired cookie/state pairs from the DB
func (s *LoginStateStore) DeleteExpired() error {
	return s.Delete("expiration_date <= NOW()").Error()
}
