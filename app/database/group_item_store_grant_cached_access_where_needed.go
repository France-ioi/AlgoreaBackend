package database

// grantCachedAccessWhereNeeded sets bCached*Access* columns to true where needed according to corresponding sCached*Access*Date columns.
// The formula is sCached*Access*Date <= NOW().
func (s *GroupItemStore) grantCachedAccessWhereNeeded() {
	listFields := map[string]string{
		"bCachedFullAccess":      "sCachedFullAccessDate",
		"bCachedPartialAccess":   "sCachedPartialAccessDate",
		"bCachedAccessSolutions": "sCachedAccessSolutionsDate",
		"bCachedGrayedAccess":    "sCachedGrayedAccessDate",
	}

	for bAccessField, sAccessDateField := range listFields {
		query := "UPDATE `groups_items` " +
			"SET `" + bAccessField + "` = true " +
			"WHERE `" + bAccessField + "` = false " +
			"AND `" + sAccessDateField + "` <= NOW()"
		mustNotBeError(s.db.Exec(query).Error)
	}
}
