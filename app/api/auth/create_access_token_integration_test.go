//go:build !unit

package auth_test

import (
	"reflect"
	"testing"
	_ "unsafe"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_createGroupFromLogin_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `groups: [{id: 1}, {id: 100}]`)
	defer func() { _ = db.Close() }()

	var nextID int64
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID", func(*database.DataStore) int64 {
		nextID++
		return nextID
	})
	defer monkey.UnpatchAll()

	dataStore := database.NewDataStore(db)
	var selfGroupID int64
	require.NoError(t, dataStore.InTransaction(func(dataStore *database.DataStore) error {
		selfGroupID = createGroupFromLogin(dataStore.Groups(), "test", &domain.CtxConfig{NonTempUsersGroupID: 100})
		return nil
	}))
	assert.Equal(t, int64(2), selfGroupID)
}

//go:linkname createGroupFromLogin github.com/France-ioi/AlgoreaBackend/v2/app/api/auth.createGroupFromLogin
func createGroupFromLogin(*database.GroupStore, string, *domain.CtxConfig) int64
