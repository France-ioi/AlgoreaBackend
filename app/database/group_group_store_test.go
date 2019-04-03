package database_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestGroupGroupStore_WhereUserIsMember(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := database.NewMockUser(1, &database.UserData{SelfGroupID: 2, OwnedGroupID: 3, DefaultLanguageID: 4})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `groups_groups` WHERE (groups_groups.idGroupChild = ? AND groups_groups.sType='direct')")).
		WithArgs(mockUser.SelfGroupID()).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := database.NewDataStore(db).GroupGroups().WhereUserIsMember(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
