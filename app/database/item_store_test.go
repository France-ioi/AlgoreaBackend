package database

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAccess(t *testing.T) {
	testCases := []struct {
		desc              string
		itemIDs           []int64
		itemAccessDetails []ItemAccessDetailsWithID
		err               error
	}{
		{
			desc:              "empty IDs",
			itemIDs:           nil,
			itemAccessDetails: nil,
			err:               nil,
		},
		{
			desc:              "empty access results",
			itemIDs:           []int64{21, 22, 23},
			itemAccessDetails: nil,
			err:               fmt.Errorf("not visible item_id 21"),
		},
		{
			desc:    "missing access result on one of the items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetails: []ItemAccessDetailsWithID{
				{ItemID: 21, ItemAccessDetails: ItemAccessDetails{CanView: "content_with_descendants"}},
				{ItemID: 22, ItemAccessDetails: ItemAccessDetails{CanView: "content_with_descendants"}},
			},
			err: fmt.Errorf("not visible item_id 23"),
		},
		{
			desc:    "no access on one of the items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetails: []ItemAccessDetailsWithID{
				{ItemID: 21, ItemAccessDetails: ItemAccessDetails{CanView: "content_with_descendants"}},
				{ItemID: 22},
				{ItemID: 23, ItemAccessDetails: ItemAccessDetails{CanView: "content_with_descendants"}},
			},
			err: fmt.Errorf("not enough perm on item_id 22"),
		},
		{
			desc:    "full access on all items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetails: []ItemAccessDetailsWithID{
				{ItemID: 21, ItemAccessDetails: ItemAccessDetails{CanView: "content_with_descendants"}},
				{ItemID: 22, ItemAccessDetails: ItemAccessDetails{CanView: "content_with_descendants"}},
				{ItemID: 23, ItemAccessDetails: ItemAccessDetails{CanView: "content_with_descendants"}},
			},
			err: nil,
		},
		{
			desc:    "content access on all but last, last is grayed",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetails: []ItemAccessDetailsWithID{
				{ItemID: 21, ItemAccessDetails: ItemAccessDetails{CanView: "content"}},
				{ItemID: 22, ItemAccessDetails: ItemAccessDetails{CanView: "content"}},
				{ItemID: 23, ItemAccessDetails: ItemAccessDetails{CanView: "info"}},
			},
			err: nil,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			err := checkAccess(tC.itemIDs, tC.itemAccessDetails)
			if err != nil {
				if tC.err != nil {
					if want, got := tC.err.Error(), err.Error(); want != got {
						t.Fatalf("Expected error to be %v, got: %v", want, got)
					}
					return
				}
				t.Fatalf("Unexpected error: %v", err)
			}
			if tC.err != nil {
				t.Fatalf("Expected error %v", tC.err)
			}
		})
	}
}

func TestItemStore_CheckSubmissionRights_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_, _, _ = NewDataStore(db).Items().CheckSubmissionRights(12, &User{GroupID: 14})
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestItemStore_ContestManagedByUser(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	dbMock.ExpectQuery(regexp.QuoteMeta("SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6)")).
		WillReturnRows(dbMock.NewRows([]string{"values"}).
			AddRow("'none','info','content','content_with_descendants','solution'"))
	dbMock.ExpectQuery(regexp.QuoteMeta("SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6)")).
		WillReturnRows(dbMock.NewRows([]string{"values"}).
			AddRow("'none','content','content_with_descendants','solution','transfer'"))
	dbMock.ExpectQuery(regexp.QuoteMeta("SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6)")).
		WillReturnRows(dbMock.NewRows([]string{"values"}).
			AddRow("'none','children','all','transfer'"))
	dbMock.ExpectQuery(regexp.QuoteMeta("SELECT items.id FROM `items` " +
		"JOIN permissions_generated ON permissions_generated.item_id = items.id " +
		"JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = permissions_generated.group_id AND " +
		"groups_ancestors_active.child_group_id = ? " +
		"WHERE (items.id = ?) AND (items.duration IS NOT NULL) " +
		"GROUP BY items.id " +
		"HAVING (MAX(permissions_generated.can_view_generated_value) >= ?) " +
		"LIMIT 1")).WillReturnRows(dbMock.NewRows([]string{"id"}).AddRow(123))
	var id int64
	err := NewDataStore(db).Items().ContestManagedByUser(123, &User{GroupID: 2}).
		PluckFirst("items.id", &id).Error()
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)
}

func TestItemStore_CanGrantViewContentOnAll_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_, _ = NewDataStore(db).Items().CanGrantViewContentOnAll(&User{GroupID: 14}, 20)
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestItemStore_CanGrantViewContentOnAll_HandlesDBErrors(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	grantViewKinds = map[string]int{"content": 3}
	defer clearAllPermissionEnums()
	dbMock.ExpectBegin()
	dbMock.ExpectQuery("").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, NewDataStore(db).InTransaction(func(store *DataStore) error {
		result, err := store.Items().CanGrantViewContentOnAll(&User{GroupID: 14}, 20)
		assert.False(t, result)
		return err
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestItemStore_AllItemsAreVisible_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_, _ = NewDataStore(db).Items().AllItemsAreVisible(&User{GroupID: 14}, 20)
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestItemStore_AllItemsAreVisible_HandlesDBErrors(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	dbMock.ExpectBegin()
	dbMock.ExpectQuery("").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, NewDataStore(db).InTransaction(func(store *DataStore) error {
		result, err := store.Items().AllItemsAreVisible(&User{GroupID: 14}, 20)
		assert.False(t, result)
		return err
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestItemStore_GetAccessDetailsForIDs_HandlesDBErrors(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	dbMock.ExpectQuery("").WillReturnError(expectedError)

	result, err := NewDataStore(db).Items().GetAccessDetailsForIDs(&User{GroupID: 14}, []int64{20})
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
