package database

const PropagationID = 1

// PropagationStore implements database operations on `propagations`
// (used for the propagation system).
type PropagationStore struct {
	*DataStore
}
