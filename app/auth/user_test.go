package auth

import (
  "context"
  "testing"

  assert_lib "github.com/stretchr/testify/assert"
)

type MockDataStore struct {}
func (s *MockDataStore) Users() UserStore {
  return &MockUserStore{}
}

type MockUserStore struct {}
func (s *MockUserStore) GetByID(userID int64, dest interface{}) error {
  u := dest.(*userData)
  u.SelfGroupID = userID+1
  return nil
}

func TestUserFromContext(t *testing.T) {
  assert := assert_lib.New(t)

  var userID int64 = 8
  ctx := context.WithValue(context.Background(), ctxUserID, userID)
  dataStore := &MockDataStore{}
  user := UserFromContext(ctx, dataStore)

  assert.EqualValues(userID, user.UserID)
  assert.NotNil(user.store)
  assert.Nil(user.data)
}

func TestSelfGroupID(t *testing.T) {
  assert := assert_lib.New(t)

  userStore := &MockUserStore{}
  user := User{42, userStore, nil}

  assert.EqualValues(43, user.SelfGroupID())
  assert.NotNil(user.data)
}
