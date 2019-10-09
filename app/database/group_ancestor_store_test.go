package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupAncestorStore_OwnedByUser(t *testing.T) {
	for _, test := range []struct {
		tableName     string
		expectedQuery string
		storeFunc     func(*DB) *GroupAncestorStore
	}{
		{
			tableName: "groups_ancestors",
			expectedQuery: "SELECT * FROM `groups_ancestors` " +
				"WHERE (`groups_ancestors`.ancestor_group_id=?) AND (NOW() < `groups_ancestors`.expires_at)",
			storeFunc: func(db *DB) *GroupAncestorStore { return NewDataStore(db).GroupAncestors() },
		},
		{
			tableName: "groups_ancestors_active",
			expectedQuery: "SELECT * FROM `groups_ancestors_active` " +
				"WHERE (`groups_ancestors_active`.ancestor_group_id=?)",
			storeFunc: func(db *DB) *GroupAncestorStore { return NewDataStore(db).ActiveGroupAncestors() },
		},
	} {
		test := test
		t.Run(test.tableName, func(t *testing.T) {
			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			mockUser := &User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(11), DefaultLanguageID: 0}
			mock.ExpectQuery(regexp.QuoteMeta(test.expectedQuery)).WithArgs(11).
				WillReturnRows(mock.NewRows([]string{"id"}))

			var result []interface{}
			err := test.storeFunc(db).OwnedByUser(mockUser).Scan(&result).Error()

			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func ptrInt64(i int64) *int64 { return &i }
