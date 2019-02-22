package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemItemStore_ChildrenOf(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const parentItemID = 123
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `items_items` WHERE (items_items.idItemParent=?)")).
		WithArgs(parentItemID).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	store := NewDataStore(db).ItemItems()
	newStore := store.ChildrenOf(parentItemID)
	assert.NotEqual(t, store, newStore)

	var result []interface{}
	err := newStore.Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
