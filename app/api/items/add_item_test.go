package items

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func Test_insertItemItems_DoesNothingWhenSpecIsEmpty(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	insertItemItems(database.NewDataStore(db), nil)
	assert.NoError(t, mock.ExpectationsWereMet())
}
