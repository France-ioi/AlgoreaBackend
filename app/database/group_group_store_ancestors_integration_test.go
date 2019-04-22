// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type groupAncestorsResultRow struct {
	GroupAncestorID int64 `gorm:"column:idGroupAncestor"`
	GroupChildID    int64 `gorm:"column:idGroupChild"`
	IsSelf          bool  `gorm:"column:bIsSelf"`
}

type groupPropagateResultRow struct {
	ID                        int64  `gorm:"column:ID"`
	AncestorsComputationState string `gorm:"column:sAncestorsComputationState"`
}

func TestGroupGroupStore_CreateNewAncestors_Concurrent(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	testhelpers.RunConcurrently(func() {
		assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
			ds.GroupGroups().CreateNewAncestors()
			return nil
		}))
	}, 30)

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("idGroupChild, idGroupAncestor").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{
		{GroupChildID: 1, GroupAncestorID: 1, IsSelf: true},

		{GroupChildID: 2, GroupAncestorID: 1, IsSelf: false},
		{GroupChildID: 2, GroupAncestorID: 2, IsSelf: true},
		{GroupChildID: 3, GroupAncestorID: 1, IsSelf: false},
		{GroupChildID: 3, GroupAncestorID: 2, IsSelf: true}, // has already been there
		{GroupChildID: 3, GroupAncestorID: 3, IsSelf: true},
		{GroupChildID: 4, GroupAncestorID: 1, IsSelf: false},
		{GroupChildID: 4, GroupAncestorID: 2, IsSelf: false},
		{GroupChildID: 4, GroupAncestorID: 3, IsSelf: false},
		{GroupChildID: 4, GroupAncestorID: 4, IsSelf: true},
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("ID").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}

func TestGroupGroupStore_CreateNewAncestors_Cyclic(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common", "group_group_store/ancestors/cyclic")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupGroups().CreateNewAncestors()
		return nil
	}))

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("idGroupChild, idGroupAncestor").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{
		{GroupChildID: 3, GroupAncestorID: 2, IsSelf: true}, // this one has already been there
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("ID").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "todo"},
		{ID: 2, AncestorsComputationState: "todo"},
		{ID: 3, AncestorsComputationState: "todo"},
		{ID: 4, AncestorsComputationState: "todo"},
	}, propagateResult)
}

func TestGroupGroupStore_CreateNewAncestors_IgnoresDoneGroups(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()

	for i := 1; i <= 4; i++ {
		assert.NoError(t, groupGroupStore.Exec(
			"INSERT INTO groups_propagate (ID, sAncestorsComputationState) VALUES (?, 'done') "+
				"ON DUPLICATE KEY UPDATE sAncestorsComputationState='done'", i).
			Error())
	}

	assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupGroups().CreateNewAncestors()
		return nil
	}))

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("idGroupChild, idGroupAncestor").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{
		{GroupChildID: 3, GroupAncestorID: 2, IsSelf: true}, // this one has already been there
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("ID").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}

func TestGroupGroupStore_CreateNewAncestors_ProcessesOnlyDirectRelationsOrAcceptedRequestsAndInvitations(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	assert.NoError(t, groupGroupStore.Exec("TRUNCATE TABLE groups_ancestors").Error())
	assert.NoError(t, groupGroupStore.Where("idGroupParent=1 AND idGroupChild=2").UpdateColumn("sType", "invitationSent").Error())
	assert.NoError(t, groupGroupStore.Where("idGroupParent=1 AND idGroupChild=3").UpdateColumn("sType", "requestSent").Error())
	assert.NoError(t, groupGroupStore.Where("idGroupParent=1 AND idGroupChild=4").UpdateColumn("sType", "invitationRefused").Error())
	assert.NoError(t, groupGroupStore.Where("idGroupParent=2 AND idGroupChild=3").UpdateColumn("sType", "requestRefused").Error())
	assert.NoError(t, groupGroupStore.Where("idGroupParent=2 AND idGroupChild=4").UpdateColumn("sType", "removed").Error())
	assert.NoError(t, groupGroupStore.Where("idGroupParent=3 AND idGroupChild=4").UpdateColumn("sType", "left").Error())

	assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupGroups().CreateNewAncestors()
		return nil
	}))

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("idGroupChild, idGroupAncestor").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{
		{GroupChildID: 1, GroupAncestorID: 1, IsSelf: true},
		{GroupChildID: 2, GroupAncestorID: 2, IsSelf: true},
		{GroupChildID: 3, GroupAncestorID: 3, IsSelf: true},
		{GroupChildID: 4, GroupAncestorID: 4, IsSelf: true},
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("ID").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}
