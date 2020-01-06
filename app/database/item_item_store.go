package database

// ItemItemStore implements database operations on `items_items`
type ItemItemStore struct {
	*DataStore
}

// ChildrenOf returns a composable query for selecting children of the given item
func (s *ItemItemStore) ChildrenOf(parentID int64) *ItemItemStore {
	return &ItemItemStore{
		NewDataStoreWithTable(s.Where("items_items.parent_item_id=?", parentID), s.tableName),
	}
}

func (s *ItemItemStore) createNewAncestors() {
	s.DataStore.createNewAncestors("items", "item")
}

// After is a "listener" that calls ItemItemStore::createNewAncestors(),
// PermissionGrantedStore::computeAllAccess() & GroupAttemptStore.ComputeAllGroupAttempts()
func (s *ItemItemStore) After() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	s.createNewAncestors()
	s.PermissionsGranted().computeAllAccess()
	return s.GroupAttempts().ComputeAllGroupAttempts()
}
