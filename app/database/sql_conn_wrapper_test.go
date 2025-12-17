package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestSQLConnWrapper_Prepare_Panics(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	require.NoError(t, db.WithFixedConnection(func(db *DB) error {
		connWrapper, ok := db.db.CommonDB().(*sqlConnWrapper)
		require.True(t, ok)
		assert.Panics(t, func() {
			_, _ = connWrapper.Prepare("SELECT 1") //nolint:sqlclosecheck // there is no statement to close as the Prepare should panic
		})
		return nil
	}))

	assert.NoError(t, mock.ExpectationsWereMet())
}
