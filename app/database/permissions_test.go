package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB_HavingMaxPermissionGreaterThan(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockDBEnumQueries(mock)
	mock.ExpectQuery("^" +
		regexp.QuoteMeta("SELECT item_id FROM `permissions_generated` HAVING (MAX(can_edit_generated_value) > ?)") + "$").
		WithArgs(3).WillReturnRows(mock.NewRows([]string{"item_id"}).AddRow(1))

	newDB := NewDataStore(db).Table("permissions_generated").
		HavingMaxPermissionGreaterThan("edit", "all")
	assert.NotEqual(t, db, newDB)

	var result []int64
	assert.NoError(t, newDB.Pluck("item_id", &result).Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
