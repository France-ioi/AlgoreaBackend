//go:build !unit

package database_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestRetryBudget_LockWaitTimeoutStopsUnderTimeBudget(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx := testhelpers.CreateTestContext()

	rawHolder := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = rawHolder.Close() }()
	rawWaiter := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = rawWaiter.Close() }()

	dbSetup, err := database.Open(ctx, rawHolder)
	require.NoError(t, err)
	defer func() { _ = dbSetup.Close() }()

	require.NoError(t, dbSetup.Exec(`
		CREATE TABLE IF NOT EXISTS lock_wait_budget_test (
			id INT PRIMARY KEY,
			val INT NOT NULL
		)`).Error())
	defer func() {
		_ = dbSetup.Exec(`DROP TABLE IF EXISTS lock_wait_budget_test`).Error()
	}()
	require.NoError(t, dbSetup.Exec(`DELETE FROM lock_wait_budget_test`).Error())
	require.NoError(t, dbSetup.Exec(`INSERT INTO lock_wait_budget_test (id, val) VALUES (1, 0)`).Error())

	holderConn, err := rawHolder.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = holderConn.Close() }()

	tx, err := holderConn.BeginTx(ctx, &sql.TxOptions{})
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	var lockedID int
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT id FROM lock_wait_budget_test WHERE id = 1 FOR UPDATE`).Scan(&lockedID))
	assert.Equal(t, 1, lockedID)

	// Short lock waits on new pool connections + a tight retry budget.
	prevParams := database.GetSessionParamsForTests()
	require.NoError(t, database.SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "1"}))
	defer func() { require.NoError(t, database.SetSessionParams(prevParams)) }()
	rawWaiter.SetMaxIdleConns(0) // drop pooled conns that still have the old timeout

	db, err := database.Open(ctx, rawWaiter)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	database.SetRetriesTimeBudgetForTests(3 * time.Second)
	defer database.ResetRetriesTimeBudgetForTests()

	started := time.Now()
	err = database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		return store.Exec("UPDATE lock_wait_budget_test SET val = 1 WHERE id = 1").Error()
	})
	elapsed := time.Since(started)

	require.Error(t, err)
	assert.Equal(t, "transaction retries time budget exceeded", err.Error())
	// budget 3s + one in-flight 1s lock wait that can start just before the deadline + slack
	assert.Less(t, elapsed, 8*time.Second,
		"should stop under the time budget rather than after retriesLimit lock waits")
	assert.GreaterOrEqual(t, elapsed, time.Second)
}

func TestMysqlConnectionWrapper_ResetSession_ReappliesConfiguredSessionParams(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx := testhelpers.CreateTestContext()
	rawDB := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = rawDB.Close() }()

	prevParams := database.GetSessionParamsForTests()
	require.NoError(t, database.SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "7"}))
	defer func() { require.NoError(t, database.SetSessionParams(prevParams)) }()

	conn, err := rawDB.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	var timeout int64
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT @@SESSION.innodb_lock_wait_timeout").Scan(&timeout))
	assert.EqualValues(t, 7, timeout)

	_, err = conn.ExecContext(ctx, "SET SESSION innodb_lock_wait_timeout = 3")
	require.NoError(t, err)

	err = conn.Raw(func(driverConn any) error {
		resetter, ok := driverConn.(driver.SessionResetter)
		require.True(t, ok)
		return resetter.ResetSession(ctx)
	})
	require.NoError(t, err)

	require.NoError(t, conn.QueryRowContext(ctx, "SELECT @@SESSION.innodb_lock_wait_timeout").Scan(&timeout))
	assert.EqualValues(t, 7, timeout)
}
