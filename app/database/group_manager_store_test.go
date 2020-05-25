package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupManagerStore_CanManage(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	groupManagerStore := NewDataStore(db).GroupManagers()

	clearAllDBEnums()
	fakeDBEnums("group_managers.can_manage",
		map[string]int{"none": 1, "memberships": 2, "memberships_and_group": 3},
		map[int]string{1: "none", 2: "memberships", 3: "memberships_and_group"})
	defer clearAllDBEnums()

	assert.Equal(t, 3, groupManagerStore.CanManageIndexByName("memberships_and_group"))
	assert.Equal(t, 2, groupManagerStore.CanManageIndexByName("memberships"))
	assert.Equal(t, 1, groupManagerStore.CanManageIndexByName("none"))
	assert.Panics(t, func() { groupManagerStore.CanManageIndexByName("unknown") })
	assert.Equal(t, "memberships_and_group", groupManagerStore.CanManageNameByIndex(3))
	assert.Equal(t, "memberships", groupManagerStore.CanManageNameByIndex(2))
	assert.Equal(t, "none", groupManagerStore.CanManageNameByIndex(1))
	assert.Panics(t, func() { groupManagerStore.CanManageNameByIndex(10) })
}
