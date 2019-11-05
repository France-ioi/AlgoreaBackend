package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionGeneratedStore_AccessRightsForItemsVisibleToUser(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{GroupID: 2, OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT item_id, MAX(can_view_generated_value) AS can_view_generated_value "+
			"FROM `permissions_generated` JOIN "+
			"( SELECT * FROM groups_ancestors_active "+
			"WHERE groups_ancestors_active.child_group_id = ? ) AS ancestors "+
			"ON ancestors.ancestor_group_id = permissions_generated.group_id GROUP BY permissions_generated.item_id "+
			"HAVING (can_view_generated_value > ?)")+"$").
		WithArgs(2, NewDataStore(db).PermissionsGranted().ViewIndexByKind("none")).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).PermissionsGenerated().AccessRightsForItemsVisibleToUser(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
