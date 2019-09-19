package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupAncestorStore_OwnedByUser(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(11), DefaultLanguageID: 0}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `groups_ancestors` WHERE (groups_ancestors.group_ancestor_id=?")).
		WithArgs(11).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).GroupAncestors().OwnedByUser(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func ptrInt64(i int64) *int64 { return &i }
