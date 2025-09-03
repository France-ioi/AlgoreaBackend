package items

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func Test_createItemItems_DoesNothingWhenSpecIsEmpty(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	insertItemItems(database.NewDataStore(db), nil)
	assert.NoError(t, mock.ExpectationsWereMet())
}
