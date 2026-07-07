package groups

import (
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func Test_groupHasPathToItemOrItemIsRoot_HasRowsErrors(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	store := database.NewDataStore(db)

	expectedErr := errors.New("db error")

	for _, test := range []struct {
		name        string
		failingCall int
	}{
		{"parent visibility query", 1},
		{"item visibility query", 2},
		{"root activity/skill query", 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			callCount := 0
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(&database.DB{}), "HasRows",
				func(_ *database.DB) (bool, error) {
					callCount++
					if callCount == test.failingCall {
						return false, expectedErr
					}
					return false, nil
				})
			defer patch.Unpatch()

			hasPath, err := groupHasPathToItemOrItemIsRoot(store, 23, 102)
			require.ErrorIs(t, err, expectedErr)
			assert.False(t, hasPath)
		})
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}
