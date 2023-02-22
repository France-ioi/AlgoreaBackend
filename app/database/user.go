package database

// User represents data associated with the user (from the `users` table)
type User struct {
	GroupID             int64
	Login               string
	LoginID             *int64
	DefaultLanguage     string
	IsAdmin             bool
	IsTempUser          bool `sql:"column:temp_user"`
	AccessGroupID       *int64
	NotificationsReadAt *Time
}

// Clone returns a deep copy of the given User structure
func (u *User) Clone() *User {
	result := *u
	if result.NotificationsReadAt != nil {
		notificationReadDateCopy := *result.NotificationsReadAt
		result.NotificationsReadAt = &notificationReadDateCopy
	}
	if result.LoginID != nil {
		loginIDCopy := *result.LoginID
		result.LoginID = &loginIDCopy
	}
	if result.AccessGroupID != nil {
		accessGroupIDCopy := *result.AccessGroupID
		result.AccessGroupID = &accessGroupIDCopy
	}
	return &result
}

// CanWatchGroupMembers checks whether user has "can_watch_members" on a group
func (u *User) CanWatchGroupMembers(store *DataStore, groupID int64) bool {
	found, err := store.ActiveGroupAncestors().ManagedByUser(u).
		Where("group_managers.can_watch_members").
		Where("groups_ancestors_active.child_group_id = ?", groupID).
		HasRows()
	mustNotBeError(err)

	return found
}
