package database

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

func TestUser_Clone(t *testing.T) {
	ts := time.Now()
	user := &User{
		Login: "login", LoginID: golang.Ptr(int64(5)), DefaultLanguage: "fr",
		IsTempUser: true, IsAdmin: true, GroupID: 2, AccessGroupID: golang.Ptr(int64(4)), NotificationsReadAt: (*Time)(&ts),
	}
	userClone := user.Clone()
	assert.NotSame(t, userClone, user)
	assert.NotSame(t, user.NotificationsReadAt, userClone.NotificationsReadAt)
	assert.Equal(t, *user.NotificationsReadAt, *userClone.NotificationsReadAt)
	assert.NotSame(t, user.LoginID, userClone.LoginID)
	assert.Equal(t, *user.LoginID, *userClone.LoginID)
	assert.NotSame(t, user.AccessGroupID, userClone.AccessGroupID)
	assert.Equal(t, *user.AccessGroupID, *userClone.AccessGroupID)
	userClone.NotificationsReadAt = nil
	userClone.LoginID = nil
	userClone.AccessGroupID = nil
	user.NotificationsReadAt = nil
	user.LoginID = nil
	user.AccessGroupID = nil
	assert.Equal(t, *user, *userClone)
	userClone = user.Clone() // clone again when pointer-type fields are nils
	assert.Nil(t, userClone.NotificationsReadAt)
	assert.Nil(t, userClone.LoginID)
	assert.Nil(t, userClone.AccessGroupID)
	assert.Equal(t, *user, *userClone)
}

func (u *User) LoadByID(dataStore *DataStore, userID int64) error {
	err := dataStore.Users().ByID(userID).
		Select(`
						users.login, users.login_id, users.is_admin, users.group_id, users.access_group_id,
						users.temp_user, users.notifications_read_at, users.default_language`).
		Take(&u).Error()
	if gorm.IsRecordNotFoundError(err) {
		u.GroupID = userID
		return nil
	}
	return err
}
