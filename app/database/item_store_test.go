package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemStore_ContestManagedByUser(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	clearAllDBEnums()
	mockDBEnumQueries(dbMock)

	dbMock.ExpectQuery(regexp.QuoteMeta("SELECT items.id FROM `items` "+
		"WHERE (items.id = ?) AND (items.duration IS NOT NULL) AND "+
		"(EXISTS(SELECT 1 FROM permissions_generated AS permissions "+
		"JOIN groups_ancestors_active AS ancestors ON ancestors.child_group_id = ? AND "+
		"ancestors.ancestor_group_id = permissions.group_id "+
		"WHERE (permissions.item_id = items.id) AND (can_view_generated_value >= ?) LIMIT 1)) "+
		"LIMIT 1")).WithArgs(int64(123), int64(2), 4).WillReturnRows(dbMock.NewRows([]string{"id"}).AddRow(123))
	var id int64
	err := NewDataStore(db).Items().ContestManagedByUser(123, &User{GroupID: 2}).
		PluckFirst("items.id", &id).Error()
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)
}
