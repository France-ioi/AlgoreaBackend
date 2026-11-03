package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestSQLDBWrapper_Prepare_Panics(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	sqlDBWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
	require.True(t, ok)
	assert.Panics(t, func() {
		_, _ = sqlDBWrapper.Prepare("SELECT 1")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
