// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestUserItemStore_ComputeAllUserItems(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "basic", wantErr: false},
	}

	db := testhelpers.SetupDBWithFixture("users_items_propagation")
	defer func() { _ = db.Close() }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := database.NewDataStore(db).UserItems()
			if err := s.ComputeAllUserItems(); (err != nil) != tt.wantErr {
				t.Errorf("UserItemStore.computeAllUserItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserItemStore_ComputeAllUserItems_Concurrent(t *testing.T) {
	rawDB, err := testhelpers.OpenRawDBConnection()
	if err != nil {
		t.Errorf("Cannot connect to the DB: %s", err)
		return
	}
	defer func() { _ = rawDB.Close() }()

	testhelpers.EmptyDB(rawDB)
	testhelpers.LoadFixture(rawDB, "users_items_propagation")

	const threadsNumber = 30
	done := make(chan bool, threadsNumber)

	db, err := database.Open(rawDB)
	if err != nil {
		t.Errorf("Cannot create a database.DB: %s", err)
		return
	}
	defer func() { _ = db.Close() }()

	for i := 0; i < threadsNumber; i++ {
		go func() {
			defer func() {
				done <- true
			}()

			s := database.NewDataStore(db)
			err := s.InTransaction(func(st *database.DataStore) error {
				return st.UserItems().ComputeAllUserItems()
			})
			assert.NoError(t, err)
		}()
	}
	for i := 0; i < threadsNumber; i++ {
		<-done
	}
}
