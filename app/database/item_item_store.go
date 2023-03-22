package database

import "time"

// ItemItemStore implements database operations on `items_items`.
type ItemItemStore struct {
	*DataStore
}

const itemsRelationsLockTimeout = 3 * time.Second

// ChildrenOf returns a composable query for selecting children of the given item.
func (s *ItemItemStore) ChildrenOf(parentID int64) *ItemItemStore {
	return &ItemItemStore{
		NewDataStoreWithTable(s.Where("items_items.parent_item_id=?", parentID), s.tableName),
	}
}

func (s *ItemItemStore) createNewAncestors() {
	s.DataStore.createNewAncestors("items", "item")
}

// After is a "listener" that calls ItemItemStore::createNewAncestors(),
// PermissionGrantedStore::computeAllAccess() and schedules a run of ResultStore.propagate().
func (s *ItemItemStore) After() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	s.createNewAncestors()
	s.PermissionsGranted().computeAllAccess()
	s.ScheduleResultsPropagation()
	return nil
}

// ContentViewPropagationNameByIndex returns the content view propagation level name with the given index from the enum.
func (s *ItemItemStore) ContentViewPropagationNameByIndex(index int) string {
	getterFunc := func() interface{} { return requireDBEnumNameByIndex("items_items.content_view_propagation", index) }
	return s.DB.getFromEnumUnderLock(getterFunc).(string)
}

// ContentViewPropagationIndexByName returns the index of the given content view propagation level in the enum.
func (s *ItemItemStore) ContentViewPropagationIndexByName(name string) int {
	getterFunc := func() interface{} { return requireDBEnumIndexByName("items_items.content_view_propagation", name) }
	return s.DB.getFromEnumUnderLock(getterFunc).(int)
}

// UpperViewLevelsPropagationNameByIndex  returns the upper-level view propagation kind name with the given index from the enum.
func (s *ItemItemStore) UpperViewLevelsPropagationNameByIndex(index int) string {
	getterFunc := func() interface{} {
		return requireDBEnumNameByIndex("items_items.upper_view_levels_propagation", index)
	}
	return s.DB.getFromEnumUnderLock(getterFunc).(string)
}

// UpperViewLevelsPropagationIndexByName returns the index of the given upper-level view propagation kind in the enum.
func (s *ItemItemStore) UpperViewLevelsPropagationIndexByName(name string) int {
	getterFunc := func() interface{} { return requireDBEnumIndexByName("items_items.upper_view_levels_propagation", name) }
	return s.DB.getFromEnumUnderLock(getterFunc).(int)
}

// WithItemsRelationsLock wraps the given function in GET_LOCK/RELEASE_LOCK
// specific for modifying relations between items.
func (s *ItemItemStore) WithItemsRelationsLock(txFunc func(*DataStore) error) error {
	return s.WithNamedLock(s.tableName, itemsRelationsLockTimeout, txFunc)
}
