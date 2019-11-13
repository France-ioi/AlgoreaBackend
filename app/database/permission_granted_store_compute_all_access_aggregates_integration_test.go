// +build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type permissionsGeneratedResultRow struct {
	GroupID          int64
	ItemID           int64
	CanViewGenerated string
}

var expectedRow14 = permissionsGeneratedResultRow{
	GroupID:          1,
	ItemID:           4,
	CanViewGenerated: "solution",
}

func TestPermissionGrantedStore_ComputeAllAccess_AggregatesContentAccess(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("permission_granted_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	permissionGrantedStore := database.NewDataStore(db).PermissionsGranted()
	permissionGeneratedStore := database.NewDataStore(db).Permissions()
	assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGeneratedStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("can_view_generated", "info").Error())
	assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.InTransaction(func(ds *database.DataStore) error {
		ds.PermissionsGranted().ComputeAllAccess()
		return nil
	}))

	assertAllPermissionsGeneratedAreDone(t, permissionGeneratedStore)

	var result []permissionsGeneratedResultRow
	assert.NoError(t, permissionGeneratedStore.Order("group_id, item_id").Scan(&result).Error())
	assertPermissionsGeneratedResultRowsEqual(t, []permissionsGeneratedResultRow{
		{
			GroupID:          1,
			ItemID:           1,
			CanViewGenerated: "content",
		},
		{
			GroupID:          1,
			ItemID:           2,
			CanViewGenerated: "content",
		},
		{
			GroupID:          1,
			ItemID:           3,
			CanViewGenerated: "content",
		},
		expectedRow14,
		{
			GroupID:          1,
			ItemID:           11,
			CanViewGenerated: "content",
		},
		{
			GroupID:          1,
			ItemID:           12,
			CanViewGenerated: "content", // content_view_propagation = 'as_content' (from 4)
		},
		{
			GroupID:          2,
			ItemID:           1,
			CanViewGenerated: "content",
		},
		{
			GroupID:          2,
			ItemID:           11,
			CanViewGenerated: "content",
		},
		{
			GroupID:          2,
			ItemID:           12,
			CanViewGenerated: "none", // content_view_propagation = 'none' (from 11)
		},
	}, result)
}

func TestPermissionGrantedStore_ComputeAllAccess_AggregatesContentAccessAsInfo(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("permission_granted_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	permissionGrantedStore := database.NewDataStore(db).PermissionsGranted()
	permissionGeneratedStore := database.NewDataStore(db).Permissions()
	assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("can_view", "content").Error())
	assert.NoError(t, permissionGrantedStore.ItemItems().Updates(map[string]interface{}{
		"content_view_propagation": "as_info",
	}).Error())
	assert.NoError(t, permissionGrantedStore.InTransaction(func(ds *database.DataStore) error {
		ds.PermissionsGranted().ComputeAllAccess()
		return nil
	}))

	assertAllPermissionsGeneratedAreDone(t, permissionGeneratedStore)

	var result []permissionsGeneratedResultRow
	assert.NoError(t, permissionGeneratedStore.Order("group_id, item_id").Scan(&result).Error())
	assertPermissionsGeneratedResultRowsEqual(t, []permissionsGeneratedResultRow{
		{
			GroupID:          1,
			ItemID:           1,
			CanViewGenerated: "content",
		},
		{
			GroupID:          1,
			ItemID:           2,
			CanViewGenerated: "content",
		},
		{
			GroupID:          1,
			ItemID:           3,
			CanViewGenerated: "content",
		},
		expectedRow14,
		{
			GroupID:          1,
			ItemID:           11,
			CanViewGenerated: "info", // since content_view_propagation = "as_info"
		},
		{
			GroupID:          1,
			ItemID:           12,
			CanViewGenerated: "info", // since content_view_propagation = "as_info"
		},
		{
			GroupID:          2,
			ItemID:           1,
			CanViewGenerated: "content",
		},
		{
			GroupID:          2,
			ItemID:           11,
			CanViewGenerated: "content",
		},
		{
			GroupID:          2,
			ItemID:           12,
			CanViewGenerated: "info", // since content_view_propagation = "as_info"
		},
	}, result)
}

func TestPermissionGrantedStore_ComputeAllAccess_AggregatesAccess(t *testing.T) {
	for _, access := range []string{"solution", "content_with_descendants"} {
		access := access
		t.Run(access, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("permission_granted_store/compute_all_access/_common")
			defer func() { _ = db.Close() }()

			permissionGrantedStore := database.NewDataStore(db).PermissionsGranted()
			permissionGeneratedStore := database.NewDataStore(db).Permissions()
			assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=1").
				UpdateColumn("can_view", access).Error())
			assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=2").
				UpdateColumn("can_view", access).Error())
			assert.NoError(t, permissionGrantedStore.Where("group_id=1 AND item_id=3").
				UpdateColumn("can_view", access).Error())
			assert.NoError(t, permissionGrantedStore.Where("group_id=2 AND item_id=1").
				UpdateColumn("can_view", access).Error())
			assert.NoError(t, permissionGrantedStore.Where("group_id=2 AND item_id=11").
				UpdateColumn("can_view", access).Error())
			assert.NoError(t, permissionGrantedStore.InTransaction(func(ds *database.DataStore) error {
				ds.PermissionsGranted().ComputeAllAccess()
				return nil
			}))

			assertAllPermissionsGeneratedAreDone(t, permissionGeneratedStore)

			var result []permissionsGeneratedResultRow
			assert.NoError(t, permissionGeneratedStore.Order("group_id, item_id").Scan(&result).Error())
			assertPermissionsGeneratedResultRowsEqual(t, []permissionsGeneratedResultRow{
				{
					GroupID:          1,
					ItemID:           1,
					CanViewGenerated: access,
				},
				{
					GroupID:          1,
					ItemID:           2,
					CanViewGenerated: access,
				},
				{
					GroupID:          1,
					ItemID:           3,
					CanViewGenerated: access,
				},
				expectedRow14,
				{
					GroupID:          1,
					ItemID:           11,
					CanViewGenerated: "content", // since content_view_propagation = "as_content"
				},
				{
					GroupID:          1,
					ItemID:           12,
					CanViewGenerated: "content", // since content_view_propagation = "as_content" (from 4)
				},
				{
					GroupID:          2,
					ItemID:           1,
					CanViewGenerated: access,
				},
				{
					GroupID:          2,
					ItemID:           11,
					CanViewGenerated: access,
				},
				{
					GroupID:          2,
					ItemID:           12,
					CanViewGenerated: "none", // since content_view_propagation = "none" (from 11)
				},
			}, result)
		})
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func assertPermissionsGeneratedResultRowsEqual(t *testing.T, expected, got []permissionsGeneratedResultRow) {
	if len(got) != len(expected) {
		assert.ElementsMatch(t, expected, got)
		return
	}

	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], got[i])
	}
}

func assertAllPermissionsGeneratedAreDone(t *testing.T, permissionGeneratedStore *database.PermissionGeneratedStore) {
	var cnt int
	assert.NoError(t, permissionGeneratedStore.Table("permissions_propagate").Count(&cnt).Error())
	assert.Zero(t, cnt, "found not done group-item pairs")
}
