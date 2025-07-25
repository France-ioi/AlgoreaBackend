package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestResultStore_ByID(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(
		"^"+regexp.QuoteMeta(
			"SELECT * FROM `results` WHERE (results.participant_id = ? AND results.attempt_id = ? AND results.item_id = ?)")+
			"$").
		WithArgs(int64(1), int64(2), int64(3)).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(123))

	var result []map[string]interface{}
	err := NewDataStore(db).Results().ByID(1, 2, 3).ScanIntoSliceOfMaps(&result).Error()
	require.NoError(t, err)
	assert.Equal(t, []map[string]interface{}{{"id": int64(123)}}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
