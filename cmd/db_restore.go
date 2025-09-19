package cmd

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func init() { //nolint:gochecknoinits // cobra suggests using init functions to add commands
	restoreCmd := &cobra.Command{
		Use:   "db-restore [environment]",
		Short: "load the last db schema",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnvToTest()
			if appenv.IsEnvProd() {
				cmd.Println("'db-restore' must not be run in 'prod' env!")
				os.Exit(1)
			}

			// load config
			dbConf, err := app.DBConfig(app.LoadConfig())
			if err != nil {
				cmd.Println("Unable to load the database config: ", err)
				os.Exit(1)
			}

			err = dropAllDBTablesWithForeignKeysChecksDisabled(dbConf)
			if err != nil {
				return err
			}

			// restore the schema
			// note: current solution is not really great as it makes some assumptions of the config :-/
			host, port, err := net.SplitHostPort(dbConf.Addr)
			if err != nil {
				host = dbConf.Addr
				port = "3306"
			}

			//nolint:gosec // we trust the config as it is filled by the user having access to the command line
			command := exec.Command(
				"mysql",
				"--host="+host,
				"--port="+port,
				"--database="+dbConf.DBName,
				"--user="+dbConf.User,
				"--password="+dbConf.Passwd,
				"--protocol=TCP",
				"-e", "source db/schema/schema.sql",
			)
			cmd.Println("mysql importing dump...")
			var output []byte
			output, err = command.CombinedOutput()
			if err != nil {
				return fmt.Errorf("command finished with error: %w\nOutput:\n%s", err, output)
			}

			// Success
			cmd.Println("DONE")

			return nil
		},
	}

	rootCmd.AddCommand(restoreCmd)
}

func dropAllDBTablesWithForeignKeysChecksDisabled(dbConf *mysql.Config) error {
	// open DB
	var db *sql.DB
	var err error
	db, err = sql.Open("mysql", dbConf.FormatDSN())
	if err != nil {
		return fmt.Errorf("unable to connect to the database: %w", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("unable to start a transaction: %w", err)
	}

	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		return fmt.Errorf("unable to query the database: %w", err)
	}

	err = dropAllDBTables(dbConf, db, tx)
	if err != nil {
		return err
	}

	// No need to restore FOREIGN_KEY_CHECKS as we close the connection.
	// Also, no need to commit the transaction as DROP TABLE is auto-committed.

	return nil
}

func dropAllDBTables(dbConf *mysql.Config, db *sql.DB, tx *sql.Tx) error {
	// remove all tables from DB
	var rows *sql.Rows
	var err error
	rows, err = db.Query(`SELECT table_name
	                      FROM   information_schema.tables
	                      WHERE  table_type   = 'BASE TABLE'
	                      AND  table_schema = ?
	                      ORDER BY table_name`, dbConf.DBName)
	if err != nil {
		return fmt.Errorf("unable to query the database: %w", err)
	}

	defer func() {
		_ = rows.Close()

		if rows.Err() != nil {
			panic(rows.Err())
		}
	}()

	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			return fmt.Errorf("unable to parse the database result: %w", err)
		}
		_, err = tx.Exec("DROP TABLE " + database.QuoteName(tableName))
		if err != nil {
			return fmt.Errorf("unable to drop table: %w", err)
		}
	}
	return nil
}
