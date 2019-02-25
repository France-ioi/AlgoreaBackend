package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemAncestorStore_DescendantsOf(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const ancestorItemID = 123
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `items_ancestors` WHERE (items_ancestors.idItemAncestor = ?)")).
		WithArgs(ancestorItemID).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).ItemAncestors().DescendantsOf(ancestorItemID).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
