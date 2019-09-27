package database

// grantCachedAccessWhereNeeded sets cached_*access* columns to true where needed according to corresponding cached*_access_*date columns.
// The formula is cached_*_access_*_date <= NOW().
func (s *GroupItemStore) grantCachedAccessWhereNeeded() {
	listFields := map[string]string{
		"cached_full_access":      "cached_full_access_since",
		"cached_partial_access":   "cached_partial_access_since",
		"cached_access_solutions": "cached_solutions_access_since",
		"cached_grayed_access":    "cached_grayed_access_since",
	}

	for accessField, accessDateField := range listFields {
		query := "UPDATE `groups_items` " +
			"SET `" + accessField + "` = true " +
			"WHERE `" + accessField + "` = false " +
			"AND `" + accessDateField + "` <= NOW()"
		mustNotBeError(s.db.Exec(query).Error)
	}
}
