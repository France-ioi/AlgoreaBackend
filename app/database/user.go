package database

// User represents data associated with the user (from the `users` table)
type User struct {
	GroupID             int64
	Login               string
	DefaultLanguage     string
	IsAdmin             bool
	IsTempUser          bool `sql:"column:temp_user"`
	AccessGroupID       *int64
	AllowSubgroups      bool
	NotificationsReadAt *Time
}

// Clone returns a deep copy of the given User structure
func (u *User) Clone() *User {
	result := *u
	if result.NotificationsReadAt != nil {
		notificationReadDateCopy := *result.NotificationsReadAt
		result.NotificationsReadAt = &notificationReadDateCopy
	}
	if result.AccessGroupID != nil {
		accessGroupIDCopy := *result.AccessGroupID
		result.AccessGroupID = &accessGroupIDCopy
	}
	return &result
}
