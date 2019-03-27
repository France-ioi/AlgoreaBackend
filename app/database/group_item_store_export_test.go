package database

func (s *GroupItemStore) GrantCachedAccessWhereNeeded() {
	s.grantCachedAccessWhereNeeded()
}

func (s *GroupItemStore) RevokeCachedAccessWhereNeeded() {
	s.revokeCachedAccessWhereNeeded()
}
