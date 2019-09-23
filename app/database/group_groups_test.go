package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB_WhereGroupRelationIsActive(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `groups_groups` " +
			"WHERE (groups_groups.type IN ('direct', 'invitationAccepted', 'requestAccepted', 'joinedByCode'))")).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := db.Table("groups_groups").WhereGroupRelationIsActive().Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
