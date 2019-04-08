package database

import (
	"errors"

	log "github.com/France-ioi/AlgoreaBackend/app/logging"
)

// User represents the context around the authenticated user making the request
// the data is loaded lazily
type User struct {
	UserID int64
	store  *UserStore
	data   *UserData
}

// UserData represents data associated with the user (from the `users` table)
type UserData struct {
	ID                int64  `sql:"column:ID"`
	Login             string `sql:"column:sLogin"`
	DefaultLanguage   string `sql:"column:sDefaultLanguage"`
	DefaultLanguageID int64  `sql:"column:idDefaultLanguage"`
	IsAdmin           bool   `sql:"column:bIsAdmin"`
	SelfGroupID       int64  `sql:"column:idGroupSelf"`
	OwnedGroupID      int64  `sql:"column:idGroupOwned"`
	AccessGroupID     int64  `sql:"column:idGroupAccess"`
	AllowSubgroups    bool   `sql:"column:allowSubgroups"`
}

var ErrUserNotFound = errors.New("user not found")

// NewUser creates a User instance
func NewUser(userID int64, userStore *UserStore, data *UserData) *User {
	return &User{UserID: userID, store: userStore, data: data}
}

func (u *User) lazyLoadData() error {
	var err error
	if u.data == nil {
		u.data = &UserData{}
		db := u.store.ByID(u.UserID).
			Joins("LEFT JOIN languages l ON (users.sDefaultLanguage = l.sCode)").
			Select("users.*, l.ID as idDefaultLanguage").
			Scan(u.data)
		if err = db.Error(); err != nil {
			u.data = nil
			log.Errorf("Unable to load user data: %s", db.Error())
		}
		if db.db.RecordNotFound() {
			return ErrUserNotFound
		}
	}
	return err
}

// SelfGroupID return the group_id of the user used for his group ownership
func (u *User) SelfGroupID() (int64, error) {
	if err := u.lazyLoadData(); err != nil {
		return 0, err
	}
	return u.data.SelfGroupID, nil
}

// DefaultLanguageID return the idDefaultLanguage of the user
func (u *User) DefaultLanguageID() (int64, error) {
	if err := u.lazyLoadData(); err != nil {
		return 0, err
	}
	return u.data.DefaultLanguageID, nil
}

// OwnedGroupID returns ID of the group that will contain groups this user manages
func (u *User) OwnedGroupID() (int64, error) {
	if err := u.lazyLoadData(); err != nil {
		return 0, err
	}
	return u.data.OwnedGroupID, nil
}

// AllowSubgroups returns if the user allowed to create subgroups
func (u *User) AllowSubgroups() (bool, error) {
	if err := u.lazyLoadData(); err != nil {
		return false, err
	}
	return u.data.AllowSubgroups, nil
}
