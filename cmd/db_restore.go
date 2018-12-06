package cmd

import (
  "database/sql"
  "fmt"
  "os"
  "os/exec"

  "github.com/France-ioi/AlgoreaBackend/app/config"
  "github.com/France-ioi/AlgoreaBackend/app/database"
  "github.com/spf13/cobra"
)

// nolint: gosec
func init() {

  var restoreCmd = &cobra.Command{
    Use:   "db-restore",
    Short: "load the last db schema",
    Run: func(cmd *cobra.Command, args []string) {
      var err error

      // load config
      var conf *config.Root
      conf, err = config.Load()
      if err != nil {
        fmt.Println("Unable to load config: ", err)
        os.Exit(1)
      }

      // open DB
      var db *database.DB
      db, err = database.DBConn(conf.Database)
      if err != nil {
        fmt.Println("Unable to connect to the database: ", err)
        os.Exit(1)
      }

      // remove all tables from DB
      var rows *sql.Rows
      rows, err = db.Raw(`SELECT CONCAT(table_schema, '.', table_name)
                          FROM   information_schema.tables
                          WHERE  table_type   = 'BASE TABLE'
                            AND  table_schema = '` + conf.Database.Connection.DBName + "'").Rows()
      if err != nil {
        fmt.Println("Unable to query the database: ", err)
        os.Exit(1)
      }
      defer rows.Close() // nolint: errcheck

      for rows.Next() {
        var tableName string
        if err = rows.Scan(&tableName); err != nil { // nolint: vetshadow
          fmt.Println("Unable to parse the database result: ", err)
          os.Exit(1)
        }
        if db.Exec("DROP TABLE " + tableName); db.Error != nil { // nolint: vetshadow
          fmt.Println("Unable to drop table: ", err)
          os.Exit(1)
        }
      }

      // restore the schema
      // note: current solution is not really great as it makes some assumptions of the config :-/
      command := exec.Command(
        "mysql",
        "-h"+conf.Database.Connection.Addr,
        "-D"+conf.Database.Connection.DBName,
        "-u"+conf.Database.Connection.User,
        "-p"+conf.Database.Connection.Passwd,
        "--protocol=TCP",
        "-e"+"source db/schema/20181024.sql",
      )
      fmt.Println("mysql importing dump...")
      err = command.Run()
      if err != nil {
        fmt.Printf("Command finished with error: %v", err)
        os.Exit(1)
      }

      // Success
      fmt.Println("DONE")
    },
  }

  rootCmd.AddCommand(restoreCmd)
}
