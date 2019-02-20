package database

// ItemAncestorStore implements database operations on `items_ancestors`
type ItemAncestorStore struct {
	*DataStore
}

func (s *ItemAncestorStore) tableName() string {
	return "items_ancestors"
}

// All creates a composable query without filtering
func (s *ItemAncestorStore) All() *ItemAncestorStore {
	return &ItemAncestorStore{&DataStore{s.Table(s.tableName())}}
}

// DescendantsOf returns a composable query for getting descendants of the given item
func (s *ItemAncestorStore) DescendantsOf(ancestorID int64) *ItemItemStore {
	return &ItemItemStore{&DataStore{
		s.Where("items_ancestors.idItemAncestor = ?", ancestorID),
	}}
}
