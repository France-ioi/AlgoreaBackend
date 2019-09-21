package database

// ItemAncestorStore implements database operations on `items_ancestors`
type ItemAncestorStore struct {
	*DataStore
}

// DescendantsOf returns a composable query for getting descendants of the given item
func (s *ItemAncestorStore) DescendantsOf(ancestorID int64) *ItemItemStore {
	return &ItemItemStore{NewDataStoreWithTable(
		s.Where("items_ancestors.ancestor_item_id = ?", ancestorID), s.tableName,
	)}
}
