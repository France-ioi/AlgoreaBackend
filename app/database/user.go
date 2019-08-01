package database

import (
	"time"
)

// User represents data associated with the user (from the `users` table)
type User struct {
	ID                   int64      `sql:"column:ID"`
	Login                string     `sql:"column:sLogin"`
	DefaultLanguage      string     `sql:"column:sDefaultLanguage"`
	DefaultLanguageID    int64      `sql:"column:idDefaultLanguage"`
	IsAdmin              bool       `sql:"column:bIsAdmin"`
	IsTempUser           bool       `sql:"column:tempUser"`
	SelfGroupID          int64      `sql:"column:idGroupSelf"`
	OwnedGroupID         int64      `sql:"column:idGroupOwned"`
	AccessGroupID        int64      `sql:"column:idGroupAccess"`
	AllowSubgroups       bool       `sql:"column:allowSubgroups"`
	NotificationReadDate *time.Time `sql:"column:sNotificationReadDate"`
}

// Clone returns a deep copy of the given User structure
func (u *User) Clone() *User {
	result := *u
	if result.NotificationReadDate != nil {
		notificationReadDateCopy := *result.NotificationReadDate
		result.NotificationReadDate = &notificationReadDateCopy
	}
	return &result
}
