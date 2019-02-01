package auth

import (
	"context"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

// User represents the context around the authenticated user making the request
// the data is loaded lazily
type User struct {
	UserID int64
	store  UserStore
	data   *userData
}
type userData struct {
	ID                int64  `sql:"column:ID"`
	Login             string `sql:"column:sLogin"`
	DefaultLanguage   string `sql:"column:sDefaultLanguage"`
	DefaultLanguageID int64  `sql:"column:idDefaultLanguage"`
	IsAdmin           bool   `sql:"column:bIsAdmin"`
	SelfGroupID       int64  `sql:"column:idGroupSelf"`
	OwnedGroupID      int64  `sql:"column:idGroupOwned"`
	AccessGroupID     int64  `sql:"column:idGroupAccess"`
}

// UserStore is an interface to the store for `users`
type UserStore interface {
	ByID(userID int64) database.DB
}

// UserFromContext creates a User context from a context set by the middleware
func UserFromContext(context context.Context, store UserStore) *User {
	userID := context.Value(ctxUserID).(int64)
	return &User{userID, store, nil}
}

func (u *User) lazyLoadData() error {
	var err error
	if u.data == nil {
		u.data = &userData{}
		db := u.store.ByID(u.UserID).
			Joins("LEFT JOIN languages l ON (users.sDefaultLanguage = l.sCode)").
			Select("users.*, l.ID as idDefaultLanguage").
			Scan(u.data)
		if err = db.Error(); err != nil {
			u.data = nil
			logging.Logger.Errorf("Unable to lazy load user data: %s", db.Error())
		}
	}
	return err
}

// SelfGroupID return the group_id of the user used for his group ownership
func (u *User) SelfGroupID() int64 {
	if u.lazyLoadData() != nil {
		return 0
	}
	return u.data.SelfGroupID
}

// DefaultLanguageID return the idDefaultLanguage of the user
func (u *User) DefaultLanguageID() int64 {
	if u.lazyLoadData() != nil {
		return 0
	}
	return u.data.DefaultLanguageID
}
