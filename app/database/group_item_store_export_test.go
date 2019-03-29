package database

import "database/sql"

func (s *GroupItemStore) GrantCachedAccessWhereNeeded() {
	s.grantCachedAccessWhereNeeded()
}

func (s *GroupItemStore) PrepareStatementsForRevokingCachedAccessWhereNeeded() []*sql.Stmt {
	return s.prepareStatementsForRevokingCachedAccessWhereNeeded()
}
