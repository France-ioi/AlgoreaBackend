package database

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestUser_Clone(t *testing.T) {
	ts := time.Now()
	user := &User{
		Login: "login", DefaultLanguage: "fr", DefaultLanguageID: 12,
		IsTempUser: true, IsAdmin: true, GroupID: 2, OwnedGroupID: ptrInt64(3), AccessGroupID: ptrInt64(4),
		AllowSubgroups: true, NotificationsReadAt: (*Time)(&ts)}
	userClone := user.Clone()
	assert.False(t, userClone == user)
	assert.False(t, user.NotificationsReadAt == userClone.NotificationsReadAt)
	assert.Equal(t, *user.NotificationsReadAt, *userClone.NotificationsReadAt)
	userClone.NotificationsReadAt = nil
	userClone.AccessGroupID = nil
	userClone.OwnedGroupID = nil
	user.NotificationsReadAt = nil
	user.AccessGroupID = nil
	user.OwnedGroupID = nil
	assert.Equal(t, *user, *userClone)
	userClone = user.Clone() // clone again when pointer-type fields are nils
	assert.Nil(t, userClone.NotificationsReadAt)
	assert.Nil(t, userClone.AccessGroupID)
	assert.Nil(t, userClone.OwnedGroupID)
	assert.Equal(t, *user, *userClone)
}

func (u *User) LoadByGroupID(dataStore *DataStore, groupID int64) error {
	err := dataStore.Users().Where("group_id = ?", groupID).
		Select(`
						users.login, users.is_admin, users.group_id, users.owned_group_id, users.access_group_id,
						users.temp_user, users.allow_subgroups, users.notifications_read_at,
						users.default_language, l.id as default_language_id`).
		Joins("LEFT JOIN languages l ON users.default_language = l.code").
		Take(&u).Error()
	if gorm.IsRecordNotFoundError(err) {
		u.GroupID = groupID
		return nil
	}
	return err
}
