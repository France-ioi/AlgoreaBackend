package database

// ComputeCached updates bCached*Access* columns according to corresponding sCached*Access*Date columns.
// The formula is sCached*Access*Date <= NOW().
func (s *GroupItemStore) ComputeCached() {
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
			"AND `" + sAccessDateField + "` IS NOT NULL AND `" + sAccessDateField + "` <= NOW()"
		mustNotBeError(s.db.Exec(query).Error)

		query = "UPDATE `groups_items` " +
			"SET `" + bAccessField + "` = false " +
			"WHERE `" + bAccessField + "` = true " +
			"AND (`" + sAccessDateField + "` IS NULL OR `" + sAccessDateField + "` > NOW())"
		mustNotBeError(s.db.Exec(query).Error)
	}
}
