package database

// GroupManagerStore implements database operations on `group_managers`
// (which stores group managers and their permissions).
type GroupManagerStore struct {
	*DataStore
}

// CanManageIndexByName returns the index of the given group manager permission from the `can_manage` enum
func (s *GroupManagerStore) CanManageIndexByName(name string) int {
	getterFunc := func() interface{} { return requireDBEnumIndexByName("group_managers.can_manage", name) }

	return s.DB.getFromEnumUnderLock(getterFunc).(int)
}

// CanManageNameByIndex returns the name of the given group manager permission from the `can_manage` enum
func (s *GroupManagerStore) CanManageNameByIndex(index int) string {
	getterFunc := func() interface{} { return requireDBEnumNameByIndex("group_managers.can_manage", index) }

	return s.DB.getFromEnumUnderLock(getterFunc).(string)
}
