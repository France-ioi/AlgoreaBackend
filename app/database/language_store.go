package database

// LanguageStore implements database operations on languages
type LanguageStore struct {
	*DataStore
}

// ByTag returns a composable query for filtering by _table_.tag
func (s *LanguageStore) ByTag(tag string) *DB {
	return s.Where(s.tableName+".tag = ?", tag)
}
