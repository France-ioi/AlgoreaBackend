package cmd

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

// nolint:gosec
func init() { // nolint:gochecknoinits,gocyclo

	var dbGenLoadCmd = &cobra.Command{
		Use:   "db-gen-load  [environment]",
		Short: "generate data for load tests",
		Run: func(cmd *cobra.Command, args []string) {
			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnv("dev")

			// load config
			conf := config.Load()
			if appenv.IsEnvProd() {
				fmt.Println("'db-gen-load' must not be run in 'prod' env!")
				os.Exit(1)
			}

			// open DB
			rawdb, err := sql.Open("mysql", conf.Database.Connection.FormatDSN())
			if err != nil {
				fmt.Println("Unable to connect to the database: ", err)
				os.Exit(1)
			}

			tablesToTruncate := []string{"groups", "groups_groups", "groups_ancestors", "groups_attempts"}
			for _, tableName := range tablesToTruncate {
				if _, err := rawdb.Exec(fmt.Sprintf("TRUNCATE %s", tableName)); err != nil {
					panic(err)
				}
			}
			_ = rawdb.Close()

			db := testhelpers.SetupDBWithFixtureString(`
			users: [{ID: 1, sLogin: owner, idGroupSelf: 21, idGroupOwned: 22}]
			groups:
				- {ID: 1, sType: Base, sName: Root, sTextId: Root}
				- {ID: 2, sType: Base, sName: RootSelf, sTextId: RootSelf}
				- {ID: 3, sType: Base, sName: RootAdmin, sTextId: RootAdmin}
				- {ID: 4, sType: Base, sName: RootTemp, sTextId: RootTemp}
				- {ID: 5, sType: Class}
				- {ID: 11, sType: Class}
				- {ID: 12, sType: Class}
				- {ID: 13, sType: Class}
				- {ID: 14, sType: Team}
				- {ID: 15, sType: Team}
				- {ID: 16, sType: Team}
				- {ID: 17, sType: Other}
				- {ID: 18, sType: Club}
				- {ID: 20, sType: Friends}
				- {ID: 21, sType: UserSelf}
				- {ID: 22, sType: UserAdmin}
			groups_groups:
				- {idGroupParent: 1, idGroupChild: 2, sType: direct, iChildOrder: 1}
				- {idGroupParent: 1, idGroupChild: 3, sType: direct, iChildOrder: 2}
				- {idGroupParent: 1, idGroupChild: 5, sType: direct}
				- {idGroupParent: 2, idGroupChild: 4, sType: direct, iChildOrder: 1}
				- {idGroupParent: 2, idGroupChild: 21, sType: direct, iChildOrder: 1}
				- {idGroupParent: 3, idGroupChild: 22, sType: direct}
				- {idGroupParent: 5, idGroupChild: 11, sType: direct}
				- {idGroupParent: 5, idGroupChild: 12, sType: direct}
				- {idGroupParent: 5, idGroupChild: 13, sType: direct}
				- {idGroupParent: 11, idGroupChild: 14, sType: direct}
				- {idGroupParent: 11, idGroupChild: 17, sType: direct}
				- {idGroupParent: 11, idGroupChild: 18, sType: direct}
				- {idGroupParent: 20, idGroupChild: 21, sType: direct}
				- {idGroupParent: 22, idGroupChild: 5, sType: direct}
			items:
				- {ID: 200, sType: Category}
				- {ID: 210, sType: Chapter}
				- {ID: 211, sType: Task}
				- {ID: 212, sType: Task}
				- {ID: 213, sType: Task}
				- {ID: 214, sType: Task}
				- {ID: 215, sType: Task}
				- {ID: 216, sType: Task}
				- {ID: 217, sType: Task}
				- {ID: 218, sType: Task}
				- {ID: 219, sType: Task}
				- {ID: 220, sType: Chapter}
				- {ID: 221, sType: Task}
				- {ID: 222, sType: Task}
				- {ID: 223, sType: Task}
				- {ID: 224, sType: Task}
				- {ID: 225, sType: Task}
				- {ID: 226, sType: Task}
				- {ID: 227, sType: Task}
				- {ID: 228, sType: Task}
				- {ID: 229, sType: Task}
				- {ID: 300, sType: Category}
				- {ID: 310, sType: Chapter}
				- {ID: 311, sType: Task}
				- {ID: 312, sType: Task}
				- {ID: 313, sType: Task}
				- {ID: 314, sType: Task}
				- {ID: 315, sType: Task}
				- {ID: 316, sType: Task}
				- {ID: 317, sType: Task}
				- {ID: 318, sType: Task}
				- {ID: 319, sType: Task}
				- {ID: 400, sType: Category}
				- {ID: 410, sType: Chapter}
				- {ID: 411, sType: Task}
				- {ID: 412, sType: Task}
				- {ID: 413, sType: Task}
				- {ID: 414, sType: Task}
				- {ID: 415, sType: Task}
				- {ID: 416, sType: Task}
				- {ID: 417, sType: Task}
				- {ID: 418, sType: Task}
				- {ID: 419, sType: Task}
			items_items:
				- {idItemParent: 200, idItemChild: 210}
				- {idItemParent: 200, idItemChild: 220}
				- {idItemParent: 210, idItemChild: 211}
				- {idItemParent: 210, idItemChild: 212}
				- {idItemParent: 210, idItemChild: 213}
				- {idItemParent: 210, idItemChild: 214}
				- {idItemParent: 210, idItemChild: 215}
				- {idItemParent: 210, idItemChild: 216}
				- {idItemParent: 210, idItemChild: 217}
				- {idItemParent: 210, idItemChild: 218}
				- {idItemParent: 210, idItemChild: 219}
				- {idItemParent: 220, idItemChild: 221}
				- {idItemParent: 220, idItemChild: 222}
				- {idItemParent: 220, idItemChild: 223}
				- {idItemParent: 220, idItemChild: 224}
				- {idItemParent: 220, idItemChild: 225}
				- {idItemParent: 220, idItemChild: 226}
				- {idItemParent: 220, idItemChild: 227}
				- {idItemParent: 220, idItemChild: 228}
				- {idItemParent: 220, idItemChild: 229}
				- {idItemParent: 300, idItemChild: 310}
				- {idItemParent: 310, idItemChild: 311}
				- {idItemParent: 310, idItemChild: 312}
				- {idItemParent: 310, idItemChild: 313}
				- {idItemParent: 310, idItemChild: 314}
				- {idItemParent: 310, idItemChild: 315}
				- {idItemParent: 310, idItemChild: 316}
				- {idItemParent: 310, idItemChild: 317}
				- {idItemParent: 310, idItemChild: 318}
				- {idItemParent: 310, idItemChild: 319}
				- {idItemParent: 400, idItemChild: 410}
				- {idItemParent: 410, idItemChild: 411}
				- {idItemParent: 410, idItemChild: 412}
				- {idItemParent: 410, idItemChild: 413}
				- {idItemParent: 410, idItemChild: 414}
				- {idItemParent: 410, idItemChild: 415}
				- {idItemParent: 410, idItemChild: 416}
				- {idItemParent: 410, idItemChild: 417}
				- {idItemParent: 410, idItemChild: 418}
				- {idItemParent: 410, idItemChild: 419}
			groups_items:
				- {idGroup: 21, idItem: 211, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 20, idItem: 212, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 213, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null, 
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 214, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 21, idItem: 215, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 216, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 217, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 20, idItem: 218, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 219, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 221, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 21, idItem: 222, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 223, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 224, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 20, idItem: 225, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 226, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 227, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 21, idItem: 228, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 229, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 311, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 20, idItem: 312, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 313, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 314, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 21, idItem: 315, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 316, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 317, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 20, idItem: 318, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 319, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 411, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 21, idItem: 412, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 413, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 414, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 20, idItem: 415, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 21, idItem: 416, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 417, sCachedFullAccessDate: null, sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}
				- {idGroup: 21, idItem: 418, sCachedFullAccessDate: null, sCachedPartialAccessDate: "2017-05-29T06:38:38Z",
					sCachedGrayedAccessDate: null}
				- {idGroup: 20, idItem: 419, sCachedFullAccessDate: "2017-05-29T06:38:38Z", sCachedPartialAccessDate: null,
					sCachedGrayedAccessDate: null}`)
			defer func() { _ = db.Close() }()

			triggersToDrop := []string{
				"before_insert_groups_attempts", "after_insert_groups_attempts",
				"before_update_groups_attempts", "before_delete_groups_attempts",
			}
			for _, triggerName := range triggersToDrop {
				fmt.Printf("\nRemoving a trigger %q. You will have to restore it manually!", triggerName)
				if err := db.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s", triggerName)).Error(); err != nil {
					panic(err)
				}
			}
			fmt.Print("\nRemoving the groups_attempts.GroupItemMinusScoreBestAnswerDateID index")
			db.Exec("ALTER TABLE groups_attempts DROP INDEX GroupItemMinusScoreBestAnswerDateID")

			limiter := make(chan bool, 5)
			barrier := make(chan bool, 100)
			done := make(chan bool)
			store := database.NewDataStore(db)

			batchSize := 250
			usersNumber := 500000
			// generate 500k students and their attempts
			for i := 0; i < usersNumber/batchSize; i++ {
				go func(i int) {
					limiter <- true
					barrier <- true
					//			usersQuery := "INSERT INTO users (ID, sLogin, idGroupSelf, idGroupOwned) VALUES "
					//			usersQueryValues := make([]string, 0, 100)

					teamsNumber := batchSize / 3
					groupsQuery := "INSERT INTO groups (ID, sType) VALUES "
					groupsQueryValues := make([]string, 0, batchSize+teamsNumber)
					groupsGroupsQuery := "INSERT INTO groups_groups (idGroupParent, idGroupChild, sType) VALUES "
					groupsGroupsQueryValues := make([]string, 0, batchSize+teamsNumber)
					groupsAttemptsQuery := "INSERT INTO groups_attempts (idGroup, idItem, sStartDate, iScore, iMinusScore, sBestAnswerDate, " +
						"nbHintsCached, nbSubmissions, bValidated, sValidationDate) VALUES "
					groupsAttemptsQueryValues := make([]string, 0, 36*batchSize)
					teams := make([]int64, teamsNumber)
					for j := 0; j < teamsNumber; j++ {
						id := int64((i * teamsNumber) + j + 100000000)
						groupsQueryValues = append(groupsQueryValues, fmt.Sprintf("(%d, 'Team')", id))
						groupsGroupsQueryValues = append(groupsGroupsQueryValues, fmt.Sprintf("(11, %d, 'direct')", id))
						teams[j] = id
					}

					for j := 0; j < batchSize; j++ {
						id := int64((i * batchSize) + j + 100)
						groupsQueryValues = append(groupsQueryValues, fmt.Sprintf("(%d, 'UserSelf'), (%d, 'UserAdmin')", id, id+1000000))
						groupsGroupsQueryValues = append(groupsGroupsQueryValues, fmt.Sprintf("(2, %d, 'direct'), (3, %d, 'direct')", id, id+1000000))

						isInTeam := rand.Float32() < 0.9
						parentGroupID := int64(11)
						if rand.Float32() < 0.5 {
							parentGroupID = 12
						}
						if isInTeam {
							parentGroupID = teams[int(rand.Float32()*float32(teamsNumber))]
						}
						groupsGroupsQueryValues = append(groupsGroupsQueryValues, fmt.Sprintf("(%d, %d, 'requestAccepted')", parentGroupID, id))

						attemptGroupID := id
						if isInTeam {
							attemptGroupID = parentGroupID
						}
						for _, itemID := range [...]int64{
							211, 212, 213, 214, 215, 216, 217, 218, 219,
							221, 222, 223, 224, 225, 226, 227, 228, 229,
							311, 312, 313, 314, 315, 316, 317, 318, 319,
							411, 412, 413, 414, 415, 416, 417, 418, 419,
						} {
							attemptsNumber := int(rand.Float32() * 2)
							for attempt := 0; attempt < attemptsNumber; attempt++ {
								score := int(rand.Float32() * 101)
								groupsAttemptsQueryValues = append(groupsAttemptsQueryValues, fmt.Sprintf(
									"(%d, %d, FROM_UNIXTIME(UNIX_TIMESTAMP('2010-04-30 14:53:27') + FLOOR(0 + (RAND() * 630720000))), "+
										"%d, %d, FROM_UNIXTIME(UNIX_TIMESTAMP('2010-04-30 14:53:27') + FLOOR(0 + (RAND() * 630720000))), "+
										"%d, %d, %d, "+
										"FROM_UNIXTIME(UNIX_TIMESTAMP('2010-04-30 14:53:27') + FLOOR(0 + (RAND() * 630720000))))",
									attemptGroupID, itemID, score, -score, int(rand.Float32()*11), int(rand.Float32()*11),
									int(rand.Float32()*2)))
							}
						}
					}
					if err := db.Exec(groupsQuery + strings.Join(groupsQueryValues, ", ")).Error(); err != nil {
						panic(err)
					}
					if err := db.Exec(groupsGroupsQuery + strings.Join(groupsGroupsQueryValues, ", ")).Error(); err != nil {
						panic(err)
					}
					if err := db.Exec(groupsAttemptsQuery + strings.Join(groupsAttemptsQueryValues, ", ")).Error(); err != nil {
						panic(err)
					}
					done <- true
				}(i)
			}
			for i := 0; i < usersNumber/batchSize; i++ {
				<-done
				if i%100 == 99 {
					// run store.GroupsGroups().createNewAncestors()
					if err := store.InTransaction(func(txStore *database.DataStore) error {
						return txStore.GroupGroups().DeleteRelation(1, 200, true)
					}); err != nil {
						panic(err)
					}
					for k := 0; k < 100; k++ {
						<-barrier
					}
				}
				<-limiter
			}
			// run store.GroupsGroups().createNewAncestors()
			if err := store.InTransaction(func(txStore *database.DataStore) error {
				return txStore.GroupGroups().DeleteRelation(1, 200, true)
			}); err != nil {
				panic(err)
			}

			fmt.Println("\nRestoring the groups_attempts.GroupItemMinusScoreBestAnswerDateID index")
			db.Exec("ALTER TABLE groups_attempts ADD INDEX GroupItemMinusScoreBestAnswerDateID (idGroup, idItem, iMinusScore, sBestAnswerDate)")

			// Success
			fmt.Println("\nDONE")
			fmt.Printf("%v\n", time.Now())
		},
	}

	rootCmd.AddCommand(dbGenLoadCmd)
}
