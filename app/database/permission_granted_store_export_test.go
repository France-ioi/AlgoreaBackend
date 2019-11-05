package database

func (s *PermissionGrantedStore) ComputeAllAccess() {
	s.computeAllAccess()
}

func (s *PermissionGrantedStore) RemovePartialAccess(groupID, itemID int64) {
	s.removePartialAccess(groupID, itemID)
}
