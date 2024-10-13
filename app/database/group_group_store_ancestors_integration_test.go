//go:build !unit

package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

type groupAncestorsResultRow struct {
	AncestorGroupID int64
	ChildGroupID    int64
	IsSelf          bool
	ExpiresAt       string
}

type groupPropagateResultRow struct {
	ID                        int64
	AncestorsComputationState string
}

var maxExpiresAt = "9999-12-31 23:59:59"

func TestGroupGroupStore_CreateNewAncestors_Concurrent(t *testing.T) {
	testhelpers.SuppressOutputIfPasses(t)

	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	testhelpers.RunConcurrently(func() {
		dataStore := database.NewDataStoreWithContext(context.Background(), db)
		assert.NoError(t, dataStore.InTransaction(func(ds *database.DataStore) error {
			ds.ScheduleGroupsAncestorsPropagation()
			return nil
		}))
	}, 30)

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("child_group_id, ancestor_group_id").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{
		{ChildGroupID: 1, AncestorGroupID: 1, IsSelf: true, ExpiresAt: maxExpiresAt},

		{ChildGroupID: 2, AncestorGroupID: 1, IsSelf: false, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 2, AncestorGroupID: 2, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 3, AncestorGroupID: 1, IsSelf: false, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 3, AncestorGroupID: 2, IsSelf: false, ExpiresAt: maxExpiresAt}, // has already been there
		{ChildGroupID: 3, AncestorGroupID: 3, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 4, AncestorGroupID: 1, IsSelf: false, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 4, AncestorGroupID: 2, IsSelf: false, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 4, AncestorGroupID: 3, IsSelf: false, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 4, AncestorGroupID: 4, IsSelf: true, ExpiresAt: maxExpiresAt},
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}

func TestGroupGroupStore_CreateNewAncestors_Cyclic(t *testing.T) {
	testhelpers.SuppressOutputIfPasses(t)

	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common", "group_group_store/ancestors/cyclic")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
		ds.ScheduleGroupsAncestorsPropagation()
		return nil
	}))

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("child_group_id, ancestor_group_id").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{ // these rows have already been there
		{ChildGroupID: 1, AncestorGroupID: 1, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 2, AncestorGroupID: 2, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 3, AncestorGroupID: 2, IsSelf: false, ExpiresAt: maxExpiresAt},
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "todo"},
		{ID: 2, AncestorsComputationState: "todo"},
		{ID: 3, AncestorsComputationState: "todo"},
		{ID: 4, AncestorsComputationState: "todo"},
	}, propagateResult)
}

func TestGroupGroupStore_CreateNewAncestors_IgnoresDoneGroups(t *testing.T) {
	testhelpers.SuppressOutputIfPasses(t)

	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()

	for i := 1; i <= 4; i++ {
		assert.NoError(t, groupGroupStore.Exec(
			"INSERT INTO groups_propagate (id, ancestors_computation_state) VALUES (?, 'done') "+
				"ON DUPLICATE KEY UPDATE ancestors_computation_state='done'", i).
			Error())
	}

	assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
		ds.ScheduleGroupsAncestorsPropagation()
		return nil
	}))

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("child_group_id, ancestor_group_id").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{ // these rows have already been there
		{ChildGroupID: 1, AncestorGroupID: 1, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 2, AncestorGroupID: 2, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 3, AncestorGroupID: 2, IsSelf: false, ExpiresAt: maxExpiresAt},
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}

func TestGroupGroupStore_CreateNewAncestors_ProcessesOnlyDirectRelationsOrAcceptedRequestsAndInvitations(t *testing.T) {
	testhelpers.SuppressOutputIfPasses(t)

	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	assert.NoError(t, groupGroupStore.Exec("TRUNCATE TABLE groups_ancestors").Error())
	assert.NoError(t, groupGroupStore.Delete("parent_group_id=1 AND child_group_id=2").Error())
	assert.NoError(t, groupGroupStore.Delete("parent_group_id=1 AND child_group_id=3").Error())
	assert.NoError(t, groupGroupStore.Delete("parent_group_id=1 AND child_group_id=4").Error())
	assert.NoError(t, groupGroupStore.Delete("parent_group_id=2 AND child_group_id=3").Error())
	assert.NoError(t, groupGroupStore.Delete("parent_group_id=2 AND child_group_id=4").Error())
	assert.NoError(t, groupGroupStore.Delete("parent_group_id=3 AND child_group_id=4").Error())

	assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
		ds.ScheduleGroupsAncestorsPropagation()
		return nil
	}))

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("child_group_id, ancestor_group_id").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{
		{ChildGroupID: 1, AncestorGroupID: 1, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 2, AncestorGroupID: 2, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 3, AncestorGroupID: 3, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 4, AncestorGroupID: 4, IsSelf: true, ExpiresAt: maxExpiresAt},
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}

func TestGroupGroupStore_CreateNewAncestors_PropagatesExpiresAt(t *testing.T) {
	testhelpers.SuppressOutputIfPasses(t)

	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	assert.NoError(t, groupGroupStore.Where("parent_group_id=1 AND child_group_id=2").
		UpdateColumn("expires_at", "3020-12-31 20:10:30").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=1 AND child_group_id=3").
		UpdateColumn("expires_at", "3019-12-31 20:10:30").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=1 AND child_group_id=4").
		UpdateColumn("expires_at", "3021-12-31 20:10:30").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=2 AND child_group_id=3").
		UpdateColumn("expires_at", "3022-12-31 20:10:30").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=2 AND child_group_id=4").
		UpdateColumn("expires_at", "3023-12-31 20:10:30").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=3 AND child_group_id=4").
		UpdateColumn("expires_at", "3024-12-31 20:10:30").Error())

	assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
		ds.ScheduleGroupsAncestorsPropagation()
		return nil
	}))

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("ancestor_group_id, child_group_id").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{
		{ChildGroupID: 1, AncestorGroupID: 1, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 2, AncestorGroupID: 1, IsSelf: false, ExpiresAt: "3020-12-31 20:10:30"},
		{ChildGroupID: 3, AncestorGroupID: 1, IsSelf: false, ExpiresAt: "3020-12-31 20:10:30"},
		{ChildGroupID: 4, AncestorGroupID: 1, IsSelf: false, ExpiresAt: "3021-12-31 20:10:30"},
		{ChildGroupID: 2, AncestorGroupID: 2, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 3, AncestorGroupID: 2, IsSelf: false, ExpiresAt: "3022-12-31 20:10:30"},
		{ChildGroupID: 4, AncestorGroupID: 2, IsSelf: false, ExpiresAt: "3023-12-31 20:10:30"},
		{ChildGroupID: 3, AncestorGroupID: 3, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 4, AncestorGroupID: 3, IsSelf: false, ExpiresAt: "3024-12-31 20:10:30"},
		{ChildGroupID: 4, AncestorGroupID: 4, IsSelf: true, ExpiresAt: maxExpiresAt},
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}

func TestGroupGroupStore_CreateNewAncestors_IgnoresExpiredRelations(t *testing.T) {
	testhelpers.SuppressOutputIfPasses(t)

	db := testhelpers.SetupDBWithFixture("group_group_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	assert.NoError(t, groupGroupStore.Where("parent_group_id=1 AND child_group_id=2").
		UpdateColumn("expires_at", "2019-05-30 20:10:30").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=1 AND child_group_id=3").
		UpdateColumn("expires_at", "2019-05-20 14:13:55").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=1 AND child_group_id=4").
		UpdateColumn("expires_at", "2019-05-15 21:13:59").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=2 AND child_group_id=3").
		UpdateColumn("expires_at", "2019-05-20 10:23:40").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=2 AND child_group_id=4").
		UpdateColumn("expires_at", "2019-05-10 12:11:45").Error())
	assert.NoError(t, groupGroupStore.Where("parent_group_id=3 AND child_group_id=4").
		UpdateColumn("expires_at", "2019-05-11 17:43:24").Error())

	assert.NoError(t, groupGroupStore.InTransaction(func(ds *database.DataStore) error {
		ds.ScheduleGroupsAncestorsPropagation()
		return nil
	}))

	var result []groupAncestorsResultRow
	assert.NoError(t, groupGroupStore.GroupAncestors().Order("ancestor_group_id, child_group_id").Scan(&result).Error())

	assert.Equal(t, []groupAncestorsResultRow{
		{ChildGroupID: 1, AncestorGroupID: 1, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 2, AncestorGroupID: 2, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 3, AncestorGroupID: 3, IsSelf: true, ExpiresAt: maxExpiresAt},
		{ChildGroupID: 4, AncestorGroupID: 4, IsSelf: true, ExpiresAt: maxExpiresAt},
	}, result)

	var propagateResult []groupPropagateResultRow
	assert.NoError(t, groupGroupStore.Table("groups_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []groupPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}
