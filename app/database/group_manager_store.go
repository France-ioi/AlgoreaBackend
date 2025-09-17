package database

// GroupManagerStore implements database operations on `group_managers`
// (which stores group managers and their permissions).
type GroupManagerStore struct {
	*DataStore
}

// CanManageIndexByName returns the index of the given group manager permission from the `can_manage` enum.
func (s *GroupManagerStore) CanManageIndexByName(name string) int {
	getterFunc := func() int { return requireDBEnumIndexByName("group_managers.can_manage", name) }

	return getFromEnumUnderLock(s.DB, getterFunc)
}

// CanManageNameByIndex returns the name of the given group manager permission from the `can_manage` enum.
func (s *GroupManagerStore) CanManageNameByIndex(index int) string {
	getterFunc := func() string { return requireDBEnumNameByIndex("group_managers.can_manage", index) }

	return getFromEnumUnderLock(s.DB, getterFunc)
}
