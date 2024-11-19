package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestItemAncestorStore_DescendantsOf(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const ancestorItemID = 123
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `items_ancestors` WHERE (items_ancestors.ancestor_item_id = ?)")).
		WithArgs(ancestorItemID).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	store := NewDataStore(db).ItemAncestors()
	newStore := store.DescendantsOf(ancestorItemID)

	assert.NotEqual(t, newStore, store)
	assert.Equal(t, "items_ancestors", newStore.DataStore.tableName)

	err := newStore.Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
