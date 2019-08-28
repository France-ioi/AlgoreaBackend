package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupStore_OwnedBy(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` " +
		"JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID " +
		"WHERE (groups_ancestors.idGroupAncestor=?)")).
		WithArgs(3).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).Groups().OwnedBy(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_TeamGroupByTeamItemAndUser(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` "+
		"JOIN groups_groups ON groups_groups.idGroupParent = groups.ID AND "+
		"groups_groups.sType"+GroupRelationIsActiveCondition+" AND "+
		"groups_groups.idGroupChild = ? "+
		"WHERE (groups.idTeamItem = ?) AND (groups.sType = 'Team') LIMIT 1")).
		WithArgs(2, 1234).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).Groups().TeamGroupByTeamItemAndUser(1234, mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_TeamGroupByItemAndUser(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` "+
		"JOIN groups_groups ON groups_groups.idGroupParent = groups.ID AND "+
		"groups_groups.sType"+GroupRelationIsActiveCondition+" AND "+
		"groups_groups.idGroupChild = ? "+
		"LEFT JOIN items_ancestors ON items_ancestors.idItemAncestor = groups.idTeamItem "+
		"WHERE (groups.sType = 'Team') AND (items_ancestors.idItemChild = ? OR groups.idTeamItem = ?) "+
		"GROUP BY groups.ID LIMIT 1")).
		WithArgs(2, 1234, 1234).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).Groups().TeamGroupByItemAndUser(1234, mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
