package database_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestGroupAncestorStore_OwnedByUser(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := auth.NewMockUser(1, 1, 11)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `groups_ancestors` WHERE (groups_ancestors.idGroupAncestor=?")).
		WithArgs(mockUser.OwnedGroupID()).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := database.NewDataStore(db).GroupAncestors().OwnedByUser(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
