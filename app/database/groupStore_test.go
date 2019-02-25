package database_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestGroupStore_OwnedBy(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := auth.NewMockUser(1, 2, 3, 4)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID WHERE (groups_ancestors.idGroupAncestor=?)")).
		WithArgs(mockUser.OwnedGroupID()).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := database.NewDataStore(db).Groups().OwnedBy(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
