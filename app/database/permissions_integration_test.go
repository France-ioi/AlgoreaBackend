//go:build !unit

package database_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

const joinsPermissionsForGroupToItemsFixture = `
		groups: [{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}]
		groups_ancestors:
			- {ancestor_group_id: 1, child_group_id: 5}
			- {ancestor_group_id: 2, child_group_id: 5, expires_at: 2019-05-30 11:00:00}
			- {ancestor_group_id: 3, child_group_id: 5}
		items: [{id: 2, default_language_tag: 2}, {id: 3, default_language_tag: 2}, {id: 4, default_language_tag: 2}]
		permissions_generated:
			- {group_id: 1, item_id: 2, can_view_generated: none, can_grant_view_generated: content, can_watch_generated: answer,
			   can_edit_generated: all, is_owner_generated: 0}
			- {group_id: 2, item_id: 2, can_view_generated: info, can_grant_view_generated: enter, can_watch_generated: answer_with_grant,
			   can_edit_generated: all_with_grant, is_owner_generated: 1}
			- {group_id: 3, item_id: 3, can_view_generated: content, can_grant_view_generated: none, can_watch_generated: answer,
				 can_edit_generated: children, is_owner_generated: 0}
			- {group_id: 4, item_id: 3, can_view_generated: content_with_descendants, can_grant_view_generated: content_with_descendants,
				 can_watch_generated: result, can_edit_generated: none, is_owner_generated: 1}
			- {group_id: 5, item_id: 2, can_view_generated: solution, can_grant_view_generated: solution, can_watch_generated: none,
				 can_edit_generated: children, is_owner_generated: 0}
			- {group_id: 5, item_id: 4, can_view_generated: none, can_grant_view_generated: enter, can_watch_generated: none,
				 can_edit_generated: children, is_owner_generated: 1}`

func TestDB_JoinsPermissionsForGroupToItems(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(joinsPermissionsForGroupToItemsFixture)
	defer func() { _ = db.Close() }()

	itemStore := database.NewDataStore(db).Items()

	for _, test := range []struct {
		ids            []int64
		expectedResult []map[string]interface{}
	}{
		{
			ids: []int64{1, 2, 3},
			expectedResult: []map[string]interface{}{
				{
					"id": int64(2), "item_id": int64(2), "can_view_generated_value": int64(5), "can_grant_view_generated_value": int64(5),
					"can_watch_generated_value": int64(3), "can_edit_generated_value": int64(3), "is_owner_generated": int64(0),
				},
				{
					"id": int64(3), "item_id": int64(3), "can_view_generated_value": int64(3), "can_grant_view_generated_value": int64(1),
					"can_watch_generated_value": int64(3), "can_edit_generated_value": int64(2), "is_owner_generated": int64(0),
				},
			},
		},
		{
			ids: []int64{1, 4, 5},
			expectedResult: []map[string]interface{}{
				{
					"id": int64(4), "item_id": int64(4), "can_view_generated_value": int64(1), "can_grant_view_generated_value": int64(2),
					"can_watch_generated_value": int64(1), "can_edit_generated_value": int64(2), "is_owner_generated": int64(1),
				},
			},
		},
	} {
		test := test
		t.Run(fmt.Sprintf("ids: %v", test.ids), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			var result []map[string]interface{}
			require.NoError(t, itemStore.Select("items.id, permissions.*").
				JoinsPermissionsForGroupToItems(5).
				Where("items.id IN (?)", test.ids).Order("items.id").
				ScanIntoSliceOfMaps(&result).Error())
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

func TestDB_JoinsPermissionsForGroupToItemsWherePermissionAtLeast(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(joinsPermissionsForGroupToItemsFixture)
	defer func() { _ = db.Close() }()

	itemStore := database.NewDataStore(db).Items()

	for _, test := range []struct {
		ids              []int64
		permissionKind   string
		neededPermission string
		expectedResult   []map[string]interface{}
	}{
		{
			ids: []int64{1, 2, 3}, permissionKind: "view", neededPermission: "info",
			expectedResult: []map[string]interface{}{
				{
					"id": int64(2), "item_id": int64(2), "can_view_generated_value": int64(5), "can_grant_view_generated_value": int64(5),
					"can_watch_generated_value": int64(3), "can_edit_generated_value": int64(3), "is_owner_generated": int64(0),
				},
				{
					"id": int64(3), "item_id": int64(3), "can_view_generated_value": int64(3), "can_grant_view_generated_value": int64(1),
					"can_watch_generated_value": int64(3), "can_edit_generated_value": int64(2), "is_owner_generated": int64(0),
				},
			},
		},
		{
			ids: []int64{1, 2, 3}, permissionKind: "view", neededPermission: "solution",
			expectedResult: []map[string]interface{}{
				{
					"id": int64(2), "item_id": int64(2), "can_view_generated_value": int64(5), "can_grant_view_generated_value": int64(5),
					"can_watch_generated_value": int64(3), "can_edit_generated_value": int64(3), "is_owner_generated": int64(0),
				},
			},
		},
		{
			ids: []int64{1, 2, 3}, permissionKind: "edit", neededPermission: "all",
			expectedResult: []map[string]interface{}{
				{
					"id": int64(2), "item_id": int64(2), "can_view_generated_value": int64(5), "can_grant_view_generated_value": int64(5),
					"can_watch_generated_value": int64(3), "can_edit_generated_value": int64(3), "is_owner_generated": int64(0),
				},
			},
		},
		{
			ids: []int64{1, 4, 5}, permissionKind: "edit", neededPermission: "children",
			expectedResult: []map[string]interface{}{
				{
					"id": int64(4), "item_id": int64(4), "can_view_generated_value": int64(1), "can_grant_view_generated_value": int64(2),
					"can_watch_generated_value": int64(1), "can_edit_generated_value": int64(2), "is_owner_generated": int64(1),
				},
			},
		},
	} {
		test := test
		t.Run(fmt.Sprintf("ids: %v, %s>=%s", test.ids, test.permissionKind, test.neededPermission), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			var result []map[string]interface{}
			require.NoError(t, itemStore.Select("items.id, permissions.*").
				JoinsPermissionsForGroupToItemsWherePermissionAtLeast(5, test.permissionKind, test.neededPermission).
				Where("items.id IN (?)", test.ids).Order("items.id").
				ScanIntoSliceOfMaps(&result).Error())
			assert.Equal(t, test.expectedResult, result)
		})
	}
}
