package database

// WherePermissionIsAtLeast returns a composable query filtered by `can_*_generated_value` >= indexOf(`permissionName`)
// depending on the given permission kind
func (conn *DB) WherePermissionIsAtLeast(permissionKind, permissionName string) *DB {
	return newDB(conn.db.
		Where(permissionColumnByKind(permissionKind)+" >= ?",
			NewDataStore(conn).PermissionsGranted().PermissionIndexByKindAndName(permissionKind, permissionName)))
}

// HavingMaxPermissionAtLeast returns a composable query filtered by `MAX(can_*_generated_value)` >= indexOf(`permissionName`)
// depending on the given permission kind
func (conn *DB) HavingMaxPermissionAtLeast(permissionKind, permissionName string) *DB {
	return newDB(conn.db.
		Having("MAX("+permissionColumnByKind(permissionKind)+") >= ?",
			NewDataStore(conn).PermissionsGranted().PermissionIndexByKindAndName(permissionKind, permissionName)))
}

// HavingMaxPermissionGreaterThan returns a composable query filtered by `MAX(can_*_generated_value)` > indexOf(`permissionName`)
// depending on the given permission kind
func (conn *DB) HavingMaxPermissionGreaterThan(permissionKind, permissionName string) *DB {
	return newDB(conn.db.
		Having("MAX("+permissionColumnByKind(permissionKind)+") > ?",
			NewDataStore(conn).PermissionsGranted().PermissionIndexByKindAndName(permissionKind, permissionName)))
}

func permissionColumnByKind(permissionKind string) string {
	return map[string]string{
		"view":       "can_view_generated_value",
		"grant_view": "can_grant_view_generated_value",
		"watch":      "can_watch_generated_value",
		"edit":       "can_edit_generated_value",
	}[permissionKind]
}
