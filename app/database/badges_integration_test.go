//go:build !unit

package database_test

import (
	"reflect"
	"testing"
	_ "unsafe"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestGroupStore_StoreBadges(t *testing.T) {
	tests := []struct {
		name                            string
		fixture                         string
		badges                          []database.Badge
		userID                          int64
		newUser                         bool
		shouldCreateBadgeGroupsForURLs  []string
		shouldMakeManagerOf             []string
		shouldMakeMemberOf              []string
		shouldCreateBadgeGroupRelations [][2]string // parent badge URL, child badge URL
		existingGroups                  []int64
		existingGroupManagers           [][2]int64 // manager, group
		existingGroupGroups             [][2]int64 // parent, child
	}{
		{
			name: "group exists, manager=false, user is a member",
			badges: []database.Badge{
				{
					URL: "abc",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def"}},
					},
				},
			},
			userID: 5,
			fixture: `
				groups: [{id: 1, text_id: abc}]
				groups_groups: [{parent_group_id: 1, child_group_id: 5}]
				groups_ancestors:
					- {ancestor_group_id: 1, child_group_id: 5}`,
			existingGroups:      []int64{1},
			existingGroupGroups: [][2]int64{{1, 5}},
		},
		{
			name: "group exists, manager=true, user is a manager",
			badges: []database.Badge{
				{
					URL:     "abc",
					Manager: true,
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def"}},
					},
				},
			},
			userID: 5,
			fixture: `
				groups: [{id: 1, text_id: abc}]
				group_managers: [{manager_id: 5, group_id: 1}]`,
			existingGroups:        []int64{1},
			existingGroupManagers: [][2]int64{{5, 1}},
		},
		{
			name: "group exists, manager=false, user is not a member",
			badges: []database.Badge{
				{
					URL: "abc",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def"}},
					},
				},
			},
			userID: 5,
			fixture: `
				groups: [{id: 1, text_id: abc}]`,
			existingGroups:     []int64{1},
			shouldMakeMemberOf: []string{"abc"},
		},
		{
			name: "group exists, manager=false, user is not a member, newUser=true",
			badges: []database.Badge{
				{
					URL: "abc",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def"}},
					},
				},
			},
			userID:  5,
			newUser: true,
			fixture: `
				groups: [{id: 1, text_id: abc}]`,
			existingGroups:     []int64{1},
			shouldMakeMemberOf: []string{"abc"},
		},
		{
			name: "group exists, manager=true, user is not a manager",
			badges: []database.Badge{
				{
					URL:     "abc",
					Manager: true,
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def"}},
					},
				},
			},
			userID: 5,
			fixture: `
				groups: [{id: 1, text_id: abc}]`,
			existingGroups:      []int64{1},
			shouldMakeManagerOf: []string{"abc"},
		},
		{
			name: "group doesn't exist, make the user a manager, parent group exists",
			badges: []database.Badge{
				{
					URL:     "abc",
					Manager: true,
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def"}},
					},
				},
			},
			userID: 5,
			fixture: `
				groups: [{id: 1, text_id: def}]`,
			existingGroups:                  []int64{1},
			shouldCreateBadgeGroupsForURLs:  []string{"abc"},
			shouldMakeManagerOf:             []string{"abc"},
			shouldCreateBadgeGroupRelations: [][2]string{{"def", "abc"}},
		},
		{
			name: "group doesn't exist, make the user a manager, no group path",
			badges: []database.Badge{
				{
					URL:     "abc",
					Manager: true,
				},
			},
			userID:                         5,
			shouldCreateBadgeGroupsForURLs: []string{"abc"},
			shouldMakeManagerOf:            []string{"abc"},
		},
		{
			name: "group doesn't exist, parent group doesn't exist, make a manager of a parent group",
			badges: []database.Badge{
				{
					URL: "abc",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def", Manager: true}},
					},
				},
			},
			userID:                          5,
			existingGroups:                  []int64{1},
			shouldCreateBadgeGroupsForURLs:  []string{"abc", "def"},
			shouldMakeMemberOf:              []string{"abc"},
			shouldMakeManagerOf:             []string{"def"},
			shouldCreateBadgeGroupRelations: [][2]string{{"def", "abc"}},
		},
		{
			name: "group doesn't exist, parent group doesn't exist, make a member of a parent group",
			badges: []database.Badge{
				{
					URL: "abc",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def"}},
					},
				},
			},
			userID:                          5,
			existingGroups:                  []int64{1},
			shouldCreateBadgeGroupsForURLs:  []string{"abc", "def"},
			shouldMakeMemberOf:              []string{"abc"},
			shouldCreateBadgeGroupRelations: [][2]string{{"def", "abc"}},
		},
		{
			name: "group path with a cycle",
			badges: []database.Badge{
				{
					URL: "abc",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "abc"}, {URL: "def"}},
					},
				},
			},
			userID:                          5,
			shouldCreateBadgeGroupsForURLs:  []string{"abc", "def"},
			shouldMakeMemberOf:              []string{"abc"},
			shouldCreateBadgeGroupRelations: [][2]string{{"def", "abc"}},
		},
		{
			name: "cannot add the user into a badge group",
			badges: []database.Badge{
				{
					URL: "abc",
				},
			},
			fixture: `
				groups: [{id: 1, text_id: abc, require_personal_info_access_approval: edit}]`,
			existingGroups: []int64{1},
			userID:         5,
		},
		{
			name: "cannot add the user into a parent badge group",
			badges: []database.Badge{
				{
					URL: "def",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "abc"}},
					},
				},
			},
			fixture: `
				groups: [{id: 1, text_id: abc, require_personal_info_access_approval: edit}]`,
			existingGroups:                  []int64{1},
			shouldCreateBadgeGroupsForURLs:  []string{"def"},
			shouldMakeMemberOf:              []string{"def"},
			shouldCreateBadgeGroupRelations: [][2]string{{"abc", "def"}},
			userID:                          5,
		},
		{
			name: "multiple badges",
			badges: []database.Badge{
				{
					URL: "abc",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "def"}, {URL: "ghi"}},
					},
				},
				{
					URL: "jkl",
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "pqr"}, {URL: "abc", Manager: true}, {URL: "mno"}},
					},
				},
			},
			userID:                          5,
			existingGroups:                  []int64{1},
			shouldCreateBadgeGroupsForURLs:  []string{"abc", "def", "ghi", "jkl", "mno"},
			shouldMakeMemberOf:              []string{"abc", "jkl"},
			shouldMakeManagerOf:             []string{"abc"},
			shouldCreateBadgeGroupRelations: [][2]string{{"ghi", "abc"}, {"def", "ghi"}, {"mno", "jkl"}, {"abc", "mno"}},
		},
		{
			name: "grand-grand-parent badge exists, grand-parent badge does not exist, parent badge exists",
			badges: []database.Badge{
				{
					URL:     "abc",
					Manager: true,
					BadgeInfo: database.BadgeInfo{
						GroupPath: []database.BadgeGroupPathElement{{URL: "jkl", Manager: true}, {URL: "def", Manager: true}, {URL: "ghi", Manager: true}},
					},
				},
			},
			fixture: `
				groups: [{id: 1, text_id: ghi}, {id: 2, text_id: jkl}]`,
			userID:                          5,
			existingGroups:                  []int64{1, 2},
			shouldCreateBadgeGroupsForURLs:  []string{"abc"},
			shouldMakeManagerOf:             []string{"abc", "ghi", "jkl"},
			shouldCreateBadgeGroupRelations: [][2]string{{"ghi", "abc"}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(`
				groups: [{id: 5}]
				users: [{group_id: 5}]` + tt.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			err := store.InTransaction(func(store *database.DataStore) error {
				return store.Groups().StoreBadges(tt.badges, tt.userID, tt.newUser)
			})
			assert.NoError(t, err)

			knownBadgeGroups := make(map[string]int64)
			expectedGroups := append([]int64{5}, tt.existingGroups...)
			for _, url := range tt.shouldCreateBadgeGroupsForURLs {
				var groupID int64
				assert.NoError(t, store.Groups().Where("text_id = ? and type='Other'", url).PluckFirst("id", &groupID).Error(),
					"a group for badge '%s' should have been created", url)
				var found bool
				found, err = store.ActiveGroupAncestors().Where("ancestor_group_id = ? and child_group_id = ?", groupID, groupID).HasRows()
				assert.NoError(t, err)
				assert.True(t, found)
				expectedGroups = append(expectedGroups, groupID)
				knownBadgeGroups[url] = groupID
			}

			expectedGroupManagers := make([]int64, 0, len(tt.existingGroupManagers)*2+len(tt.shouldMakeManagerOf)*2)
			for _, idsPair := range tt.existingGroupManagers {
				managerGroupID, groupID := idsPair[0], idsPair[1]
				expectedGroupManagers = append(expectedGroupManagers, managerGroupID, groupID)
			}
			for _, url := range tt.shouldMakeManagerOf {
				groupID := getGroupIDByBadgeURL(store, url, knownBadgeGroups)
				var found bool
				found, err = store.GroupManagers().
					Where("group_id = ?", groupID).
					Where("manager_id = ?", tt.userID).
					Where("can_manage = 'memberships'").
					Where("can_grant_group_access").
					Where("can_watch_members").HasRows()
				assert.NoError(t, err)
				assert.True(t, found, "the user should have become a manager of badge '%s'", url)
				expectedGroupManagers = append(expectedGroupManagers, tt.userID, groupID)
			}

			expectedGroupsGroups := make(
				[]int64,
				0,
				len(tt.existingGroupGroups)*2+len(tt.shouldMakeMemberOf)*2+len(tt.shouldCreateBadgeGroupRelations)*2,
			)
			for _, idsPair := range tt.existingGroupGroups {
				parentGroupID, childGroupID := idsPair[0], idsPair[1]
				expectedGroupsGroups = append(expectedGroupsGroups, parentGroupID, childGroupID)
			}
			for _, url := range tt.shouldMakeMemberOf {
				groupID := getGroupIDByBadgeURL(store, url, knownBadgeGroups)
				var found bool
				found, err = store.ActiveGroupGroups().
					Where("parent_group_id = ?", groupID).
					Where("child_group_id = ?", tt.userID).HasRows()
				assert.NoError(t, err)
				assert.True(t, found, "the user should have become a member of badge '%s'", url)
				found, err = store.ActiveGroupAncestors().
					Where("ancestor_group_id = ?", groupID).
					Where("child_group_id = ?", tt.userID).HasRows()
				assert.NoError(t, err)
				assert.True(t, found, "the user should have become a descendant of badge '%s'", url)
				expectedGroupsGroups = append(expectedGroupsGroups, groupID, tt.userID)
			}
			for _, urlPair := range tt.shouldCreateBadgeGroupRelations {
				parentBadgeURL, childBadgeURL := urlPair[0], urlPair[1]
				parentGroupID := getGroupIDByBadgeURL(store, parentBadgeURL, knownBadgeGroups)
				childGroupID := getGroupIDByBadgeURL(store, childBadgeURL, knownBadgeGroups)
				var found bool
				found, err = store.ActiveGroupGroups().
					Where("parent_group_id = ?", parentGroupID).
					Where("child_group_id = ?", childGroupID).HasRows()
				assert.NoError(t, err)
				assert.True(t, found, "the badge '%s' should have become a subgroup of badge '%s'", childBadgeURL, parentBadgeURL)
				found, err = store.ActiveGroupAncestors().
					Where("ancestor_group_id = ?", parentGroupID).
					Where("child_group_id = ?", childGroupID).HasRows()
				assert.NoError(t, err)
				assert.True(t, found, "the badge '%s' should have become a descendant of badge '%s'", childBadgeURL, parentBadgeURL)
				expectedGroupsGroups = append(expectedGroupsGroups, parentGroupID, childGroupID)
			}

			found, err := store.Groups().Where("id NOT IN(?)", expectedGroups).HasRows()
			assert.NoError(t, err)
			assert.False(t, found, "some unexpected groups have been created")

			groupManagersQuery := store.GroupManagers().DB
			for index := 0; index < len(expectedGroupManagers); index += 2 {
				groupManagersQuery = groupManagersQuery.Where("NOT (manager_id = ? AND group_id = ?)",
					expectedGroupManagers[index], expectedGroupManagers[index+1])
			}
			found, err = groupManagersQuery.HasRows()
			assert.NoError(t, err)
			assert.False(t, found, "some unexpected group_managers have been created")

			groupsGroupsQuery := store.GroupGroups().DB
			for index := 0; index < len(expectedGroupsGroups); index += 2 {
				groupsGroupsQuery = groupsGroupsQuery.Where("NOT (parent_group_id = ? AND child_group_id = ?)",
					expectedGroupsGroups[index], expectedGroupsGroups[index+1])
			}
			found, err = groupsGroupsQuery.HasRows()
			assert.NoError(t, err)
			assert.False(t, found, "some unexpected groups_groups have been created")
		})
	}
}

func TestGroupStore_StoreBadge_PropagatesResults(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
				groups: [{id: 1, text_id: badge_url}, {id: 5}, {id: 6}]
				users: [{group_id: 5, login: john}, {group_id: 6, login: jane}]
				groups_groups: [{parent_group_id: 1, child_group_id: 5}]
				groups_ancestors:
					- {ancestor_group_id: 1, child_group_id: 5}
				items: [{id: 100, default_language_tag: fr}, {id: 101, default_language_tag: fr}]
				items_items: [{parent_item_id: 100, child_item_id: 101, child_order: 1}]
				items_ancestors: [{ancestor_item_id: 100, child_item_id: 101}]
				permissions_generated:
					- {group_id: 1, item_id: 100, can_view_generated: content}
					- {group_id: 1, item_id: 101, can_view_generated: content}
				attempts: [{participant_id: 5, id: 1}, {participant_id: 6, id: 1}]
				results:
					- {participant_id: 5, item_id: 101, attempt_id: 1, score_computed: 100}
					- {participant_id: 6, item_id: 101, attempt_id: 1, score_computed: 10}
`)
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)
	err := store.InTransaction(func(store *database.DataStore) error {
		return store.Groups().StoreBadges([]database.Badge{{URL: "badge_url"}}, 6, false)
	})
	assert.NoError(t, err)

	found, err := store.Table("results_propagate").HasRows()
	assert.NoError(t, err)
	assert.False(t, found)

	var score float32
	assert.NoError(t, store.Results().ByID(6, 1, 100).
		PluckFirst("score_computed", &score).Error())
	assert.Equal(t, float32(10), score)
}

func getGroupIDByBadgeURL(store *database.DataStore, url string, knownBadgeGroups map[string]int64) int64 {
	if id, ok := knownBadgeGroups[url]; ok {
		return id
	}
	var id int64
	err := store.Groups().Where("text_id = ?", url).PluckFirst("id", &id).Error()
	if err != nil {
		panic(err)
	}
	knownBadgeGroups[url] = id
	return id
}

func TestGroupStore_createBadgeGroup_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1, text_id: badge_url}]`)
	defer func() { _ = db.Close() }()
	groupStore := database.NewDataStore(db).Groups()

	var nextID int64
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID", func(_ *database.DataStore) int64 {
		nextID++
		return nextID
	})
	defer monkey.UnpatchAll()

	newID := groupStoreCreateBadgeGroup(groupStore, "url", "name")
	assert.Equal(t, int64(2), newID)
}

//go:linkname groupStoreCreateBadgeGroup github.com/France-ioi/AlgoreaBackend/v2/app/database.(*GroupStore).createBadgeGroup
func groupStoreCreateBadgeGroup(store *database.GroupStore, badgeURL, badgeName string) int64
