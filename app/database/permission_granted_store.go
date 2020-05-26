package database

// PermissionGrantedStore implements database operations on `permissions_granted`
type PermissionGrantedStore struct {
	*DataStore
}

// After is a "listener" that calls PermissionGrantedStore::computeAllAccess()
func (s *PermissionGrantedStore) After() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	s.computeAllAccess()
	return nil
}

// PermissionIndexByKindAndName returns the index of the given permission in the enum
func (s *PermissionGrantedStore) PermissionIndexByKindAndName(kind, name string) int {
	getterFunc := func() interface{} { return requireDBEnumIndexByName("permissions_granted.can_"+kind, name) }
	return s.DB.getFromEnumUnderLock(getterFunc).(int)
}

// ViewIndexByName returns the index of the given view kind in the 'can_view' enum
func (s *PermissionGrantedStore) ViewIndexByName(name string) int {
	return s.PermissionIndexByKindAndName("view", name)
}

// PermissionNameByKindAndIndex returns the permission name of the given kind with the given index from the enum
func (s *PermissionGrantedStore) PermissionNameByKindAndIndex(kind string, index int) string {
	getterFunc := func() interface{} { return requireDBEnumNameByIndex("permissions_granted.can_"+kind, index) }
	return s.DB.getFromEnumUnderLock(getterFunc).(string)
}

// ViewNameByIndex returns the view permission name with the given index from the 'can_view' enum
func (s *PermissionGrantedStore) ViewNameByIndex(index int) string {
	return s.PermissionNameByKindAndIndex("view", index)
}

// GrantViewIndexByName returns the index of the given "grant view" permission name in the 'can_grant_view' enum
func (s *PermissionGrantedStore) GrantViewIndexByName(name string) int {
	return s.PermissionIndexByKindAndName("grant_view", name)
}

// EditIndexByName returns the index of the given "edit" permission name in the 'can_edit' enum
func (s *PermissionGrantedStore) EditIndexByName(name string) int {
	return s.PermissionIndexByKindAndName("edit", name)
}
