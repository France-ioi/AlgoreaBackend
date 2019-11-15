package database

func (s *PermissionGrantedStore) ComputeAllAccess() {
	s.computeAllAccess()
}

func (s *PermissionGrantedStore) RemoveContentAccess(groupID, itemID int64) {
	s.removeContentAccess(groupID, itemID)
}
