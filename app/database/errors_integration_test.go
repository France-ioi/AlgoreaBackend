//go:build !unit

package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database/mysqldb"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestIsKindOfErrors(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext())
	defer func() { _ = db.Close() }()

	require.NoError(t, db.Exec("CREATE TABLE test_fk_parent (id INT PRIMARY KEY)").Error())
	defer func() {
		_ = db.Exec("DROP TABLE test_fk_parent").Error()
	}()
	require.NoError(t,
		db.Exec(`
			CREATE TABLE test_fk_child (id INT PRIMARY KEY, parent_id INT, FOREIGN KEY (parent_id)
				REFERENCES test_fk_parent(id))`).Error())
	defer func() {
		_ = db.Exec("DROP TABLE test_fk_child").Error()
	}()
	require.NoError(t, db.Exec("INSERT INTO test_fk_parent (id) VALUES (1)").Error())
	require.NoError(t, db.Exec("INSERT INTO test_fk_child (id, parent_id) VALUES (1, 1)").Error())

	require.NoError(t, db.Exec("CREATE USER 'another_user'@'%' IDENTIFIED BY 'another_password'").Error())
	defer func() {
		_ = db.Exec("DROP USER 'another_user'@'%'").Error()
	}()
	require.NoError(t, db.Exec("GRANT SELECT, DELETE ON test_fk_parent TO 'another_user'@'%'").Error())
	db1, err := database.Open(db.GetContext(), openRawDBConnectionForAnotherUser(db.GetContext(), "another_user", "another_password"))
	require.NoError(t, err)
	defer func() { _ = db1.Close() }()

	t.Run("IsKindOfRowIsReferencedError", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		defer func() {
			require.NoError(t, db.Exec("REVOKE SELECT, DELETE ON test_fk_parent FROM 'another_user'@'%'").Error())
		}()
		t.Run("with REFERENCES privilege", func(t *testing.T) {
			testoutput.SuppressIfPasses(t)
			testIsKindOfRowIsReferencedErrorWithDB(t, db, mysqldb.RowIsReferenced2)
		})
		t.Run("without REFERENCES privilege", func(t *testing.T) {
			testoutput.SuppressIfPasses(t)
			testIsKindOfRowIsReferencedErrorWithDB(t, db1, mysqldb.RowIsReferenced)
		})
	})
	t.Run("IsKindOfNoReferencedRowError", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		require.NoError(t, db.Exec("GRANT INSERT ON test_fk_child TO 'another_user'@'%'").Error())
		defer func() {
			require.NoError(t, db.Exec("REVOKE INSERT ON test_fk_child FROM 'another_user'@'%'").Error())
		}()
		t.Run("with REFERENCES privilege", func(t *testing.T) {
			testoutput.SuppressIfPasses(t)
			testIsKindOfNoReferencedRowErrorWithDB(t, db, mysqldb.NoReferencedRow2)
		})
		t.Run("without REFERENCES privilege", func(t *testing.T) {
			testoutput.SuppressIfPasses(t)
			testIsKindOfNoReferencedRowErrorWithDB(t, db1, mysqldb.NoReferencedRow)
		})
	})
}

func testIsKindOfRowIsReferencedErrorWithDB(t *testing.T, db *database.DB, expectedErrorNumber mysqldb.MysqlErrorNumber) {
	t.Helper()

	// Attempt to delete a parent row that is referenced by a child row
	err := db.Exec("DELETE FROM test_fk_parent WHERE id = 1").Error()
	require.Error(t, err)

	require.True(t, mysqldb.IsMysqlError(err, expectedErrorNumber),
		"expected error number %d for error: %v", expectedErrorNumber, err)
	assert.True(t, database.IsKindOfRowIsReferencedError(err))
}

func testIsKindOfNoReferencedRowErrorWithDB(t *testing.T, db *database.DB, expectedErrorNumber mysqldb.MysqlErrorNumber) {
	t.Helper()

	// Attempt to insert a child row with a non-existing parent_id
	err := db.Exec("INSERT INTO test_fk_child (id, parent_id) VALUES (2, 999)").Error()
	require.Error(t, err)

	require.True(t, mysqldb.IsMysqlError(err, expectedErrorNumber),
		"expected error number %d for error: %v", expectedErrorNumber, err)
	require.True(t, database.IsKindOfNoReferencedRowError(err))
}

func openRawDBConnectionForAnotherUser(ctx context.Context, userName, userPassword string) *sql.DB {
	// needs actual config for connection to DB
	config := testhelpers.GetConfigFromContext(ctx)
	dbConfig, _ := app.DBConfig(config)
	if dbConfig.Params == nil {
		dbConfig.Params = make(map[string]string, 1)
	}
	dbConfig.Params["charset"] = "utf8mb4"

	dbConfig.User = userName
	dbConfig.Passwd = userPassword

	logger := logging.LoggerFromContext(ctx)
	rawDB, err := database.OpenRawDBConnection(dbConfig.FormatDSN(), logger.IsRawSQLQueriesLoggingEnabled())
	if err != nil {
		panic(err)
	}
	return rawDB
}
