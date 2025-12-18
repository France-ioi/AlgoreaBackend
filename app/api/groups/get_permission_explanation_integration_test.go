//go:build !unit

package groups_test

import (
	"testing"
	_ "unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_insertGrantedPermissionsToBeExplained(t *testing.T) {
	tests := []struct {
		name      string
		dbFixture string
		itemID    int64
		groupID   int64
		expected  []map[string]interface{}
	}{
		{
			name:    "inserts non-empty permissions with different source groups and different origins",
			groupID: 1, itemID: 2,
			dbFixture: `
				groups: [{id: 1, type: Class}, {id: 3, type: Class}, {id: 4, type: Class}, {id: 5, type: Class},
				         {id: 6, type: Class}, {id: 7, type: Class}]
				items: [{id: 2, default_language_tag: en}]
				permissions_granted:
					- {group_id: 1, item_id: 2, source_group_id: 3, origin: group_membership, can_view: info}
					- {group_id: 1, item_id: 2, source_group_id: 4, origin: group_membership, can_view: content}
					- {group_id: 1, item_id: 2, source_group_id: 4, origin: item_unlocking, can_view: content_with_descendants}
					- {group_id: 1, item_id: 2, source_group_id: 4, origin: self, can_view: solution}
					- {group_id: 1, item_id: 2, source_group_id: 4, origin: other, can_grant_view: enter}
					- {group_id: 1, item_id: 2, source_group_id: 5, origin: group_membership, can_grant_view: content}
					- {group_id: 1, item_id: 2, source_group_id: 5, origin: item_unlocking, can_grant_view: content_with_descendants}
					- {group_id: 1, item_id: 2, source_group_id: 5, origin: self, can_grant_view: solution}
					- {group_id: 1, item_id: 2, source_group_id: 5, origin: other, can_grant_view: solution_with_grant}
					- {group_id: 1, item_id: 2, source_group_id: 6, origin: group_membership, can_watch: result}
					- {group_id: 1, item_id: 2, source_group_id: 6, origin: item_unlocking, can_watch: answer}
					- {group_id: 1, item_id: 2, source_group_id: 6, origin: self, can_watch: answer_with_grant}
					- {group_id: 1, item_id: 2, source_group_id: 7, origin: group_membership, can_edit: children}
					- {group_id: 1, item_id: 2, source_group_id: 7, origin: item_unlocking, can_edit: all}
					- {group_id: 1, item_id: 2, source_group_id: 7, origin: self, can_edit: all_with_grant}
					- {group_id: 1, item_id: 2, source_group_id: 7, origin: other, is_owner: true}
			`,
			expected: []map[string]interface{}{
				{
					"group_id": "1|2|3|group_membership", "item_id": int64(2), "source_group_id": int64(3), "origin": "group_membership",
					"can_view": "info", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|4|group_membership", "item_id": int64(2), "source_group_id": int64(4), "origin": "group_membership",
					"can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|4|item_unlocking", "item_id": int64(2), "source_group_id": int64(4), "origin": "item_unlocking",
					"can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|4|self", "item_id": int64(2), "source_group_id": int64(4), "origin": "self",
					"can_view": "solution", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|4|other", "item_id": int64(2), "source_group_id": int64(4), "origin": "other",
					"can_view": "none", "can_grant_view": "enter", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|5|group_membership", "item_id": int64(2), "source_group_id": int64(5), "origin": "group_membership",
					"can_view": "none", "can_grant_view": "content", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|5|item_unlocking", "item_id": int64(2), "source_group_id": int64(5), "origin": "item_unlocking",
					"can_view": "none", "can_grant_view": "content_with_descendants", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|5|self", "item_id": int64(2), "source_group_id": int64(5), "origin": "self",
					"can_view": "none", "can_grant_view": "solution", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|5|other", "item_id": int64(2), "source_group_id": int64(5), "origin": "other",
					"can_view": "none", "can_grant_view": "solution_with_grant", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|6|group_membership", "item_id": int64(2), "source_group_id": int64(6), "origin": "group_membership",
					"can_view": "none", "can_grant_view": "none", "can_watch": "result", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|6|item_unlocking", "item_id": int64(2), "source_group_id": int64(6), "origin": "item_unlocking",
					"can_view": "none", "can_grant_view": "none", "can_watch": "answer", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|6|self", "item_id": int64(2), "source_group_id": int64(6), "origin": "self",
					"can_view": "none", "can_grant_view": "none", "can_watch": "answer_with_grant", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|7|group_membership", "item_id": int64(2), "source_group_id": int64(7), "origin": "group_membership",
					"can_view": "none", "can_grant_view": "none", "can_watch": "none", "can_edit": "children", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|7|item_unlocking", "item_id": int64(2), "source_group_id": int64(7), "origin": "item_unlocking",
					"can_view": "none", "can_grant_view": "none", "can_watch": "none", "can_edit": "all", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|7|self", "item_id": int64(2), "source_group_id": int64(7), "origin": "self",
					"can_view": "none", "can_grant_view": "none", "can_watch": "none", "can_edit": "all_with_grant", "is_owner": int64(0),
				},
				{
					"group_id": "1|2|7|other", "item_id": int64(2), "source_group_id": int64(7), "origin": "other",
					"can_view": "none", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(1),
				},
			},
		},
		{
			name:    "inserts non-empty permissions for ancestor groups",
			groupID: 1, itemID: 2,
			dbFixture: `
				groups: [{id: 1, type: Class}, {id: 3, type: Class}, {id: 4, type: Class}, {id: 5, type: Team},
				         {id: 6, type: Class}, {id: 7, type: Class}]
				groups_groups:
					- {parent_group_id: 5, child_group_id: 1}
					- {parent_group_id: 6, child_group_id: 1}
					- {parent_group_id: 7, child_group_id: 6}
				items: [{id: 2, default_language_tag: en}]
				permissions_granted:
					- {group_id: 5, item_id: 2, source_group_id: 3, origin: group_membership, can_view: info}
					- {group_id: 6, item_id: 2, source_group_id: 4, origin: group_membership, can_view: content}
					- {group_id: 7, item_id: 2, source_group_id: 4, origin: item_unlocking, can_view: content_with_descendants}
			`,
			expected: []map[string]interface{}{
				{
					"group_id": "6|2|4|group_membership", "item_id": int64(2), "source_group_id": int64(4), "origin": "group_membership",
					"can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "7|2|4|item_unlocking", "item_id": int64(2), "source_group_id": int64(4), "origin": "item_unlocking",
					"can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
			},
		},
		{
			name:    "inserts non-empty permissions for ancestor items",
			groupID: 1, itemID: 2,
			dbFixture: `
				groups: [{id: 1, type: Class}, {id: 3, type: Class}, {id: 4, type: Class}, {id: 5, type: Class},
				         {id: 6, type: Class}, {id: 7, type: Class}]
				items: [{id: 2, default_language_tag: en}, {id: 3, default_language_tag: en}, {id: 4, default_language_tag: en}]
				items_items:
					- {parent_item_id: 3, child_item_id: 2, child_order: 1}
					- {parent_item_id: 4, child_item_id: 3, child_order: 1}
				permissions_granted:
					- {group_id: 1, item_id: 3, source_group_id: 3, origin: group_membership, can_view: info}
					- {group_id: 1, item_id: 4, source_group_id: 4, origin: group_membership, can_view: content}
			`,
			expected: []map[string]interface{}{
				{
					"group_id": "1|3|3|group_membership", "item_id": int64(3), "source_group_id": int64(3), "origin": "group_membership",
					"can_view": "info", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
				{
					"group_id": "1|4|4|group_membership", "item_id": int64(4), "source_group_id": int64(4), "origin": "group_membership",
					"can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": int64(0),
				},
			},
		},
		{
			name:    "skips empty permissions",
			groupID: 1, itemID: 2,
			dbFixture: `
				groups: [{id: 1, type: Class}, {id: 3, type: Class}, {id: 4, type: Class}, {id: 5, type: Class},
				         {id: 6, type: Class}, {id: 7, type: Class}]
				groups_groups:
					- {parent_group_id: 5, child_group_id: 1}
					- {parent_group_id: 6, child_group_id: 1}
					- {parent_group_id: 7, child_group_id: 6}
				items: [{id: 2, default_language_tag: en}, {id: 3, default_language_tag: en}, {id: 4, default_language_tag: en}]
				items_items:
					- {parent_item_id: 3, child_item_id: 2, child_order: 1}
					- {parent_item_id: 4, child_item_id: 3, child_order: 1}
				permissions_granted:
					- {group_id: 1, item_id: 3, source_group_id: 3, origin: group_membership}
					- {group_id: 1, item_id: 4, source_group_id: 4, origin: group_membership}
					- {group_id: 5, item_id: 2, source_group_id: 5, origin: group_membership}
					- {group_id: 6, item_id: 2, source_group_id: 6, origin: group_membership}
					- {group_id: 7, item_id: 2, source_group_id: 7, origin: group_membership}
			`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), tt.dbFixture)
			defer func() { _ = db.Close() }()
			require.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				require.NoError(t, store.GroupGroups().CreateNewAncestors())
				require.NoError(t, store.ItemItems().CreateNewAncestors())
				return nil
			}))

			require.NoError(t, db.WithFixedConnection(func(db *database.DB) error {
				cleanupFunc, err := database.NewDataStore(db).PermissionsGranted().CreateTemporaryTablesForPermissionsExplanation()
				defer cleanupFunc()
				require.NoError(t, err)

				insertGrantedPermissionsToBeExplained(db, tt.itemID, tt.groupID)

				var result []map[string]interface{}
				require.NoError(t, db.Table("permissions_granted_exp").Select(`
					group_id, item_id, source_group_id, origin, can_view, can_grant_view, can_watch, can_edit, is_owner
				`).Order("CAST(SUBSTRING_INDEX(group_id, '|', 1) AS SIGNED), item_id, source_group_id, origin").ScanIntoSliceOfMaps(&result).Error())

				assert.Equal(t, tt.expected, result)
				return nil
			}))
		})
	}
}

//go:linkname insertGrantedPermissionsToBeExplained github.com/France-ioi/AlgoreaBackend/v2/app/api/groups.insertGrantedPermissionsToBeExplained
func insertGrantedPermissionsToBeExplained(db *database.DB, itemID, groupID int64)
