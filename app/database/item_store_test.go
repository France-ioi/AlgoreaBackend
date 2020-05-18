package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemStore_ContestManagedByUser(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	clearAllPermissionEnums()
	mockPermissionEnumQueries(dbMock)

	dbMock.ExpectQuery(regexp.QuoteMeta("SELECT items.id FROM `items` " +
		"JOIN permissions_generated ON permissions_generated.item_id = items.id " +
		"JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = permissions_generated.group_id AND " +
		"groups_ancestors_active.child_group_id = ? " +
		"WHERE (items.id = ?) AND (items.duration IS NOT NULL) " +
		"GROUP BY items.id " +
		"HAVING (MAX(can_view_generated_value) >= ?) " +
		"LIMIT 1")).WillReturnRows(dbMock.NewRows([]string{"id"}).AddRow(123))
	var id int64
	err := NewDataStore(db).Items().ContestManagedByUser(123, &User{GroupID: 2}).
		PluckFirst("items.id", &id).Error()
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)
}
