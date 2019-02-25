package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB_WhereUsersAreDescendantsOfGroup(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const groupID = 123
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `users`.* FROM `users` JOIN groups_ancestors ON groups_ancestors.idGroupChild=users.idGroupSelf WHERE (groups_ancestors.idGroupAncestor = ?)")).
		WithArgs(groupID).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := db.Table("users").WhereUsersAreDescendantsOfGroup(groupID).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
