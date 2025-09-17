package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestSQLTxWrapper_Prepare_Panics(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	assert.NoError(t, db.inTransaction(func(db *DB) error {
		sqlTxWrapper, ok := db.db.CommonDB().(*sqlTxWrapper)
		require.True(t, ok)
		assert.Panics(t, func() {
			_, _ = sqlTxWrapper.Prepare("SELECT 1") //nolint:sqlclosecheck // there is no statement to close as the Prepare should panic
		})
		return nil
	}))

	assert.NoError(t, mock.ExpectationsWereMet())
}
