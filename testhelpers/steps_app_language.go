// +build !prod

package testhelpers

import (
	"strconv"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/rand"
)

// getParametersMap parses parameters in format key1=val1,key2=val2,... into a map.
func getParametersMap(parameters string) map[string]interface{} {
	parametersMap := make(map[string]interface{})
	arrayParameters := strings.Split(parameters, ",")
	for _, paramKeyValue := range arrayParameters {
		keyVal := strings.Split(paramKeyValue, "=")
		parametersMap[keyVal[0]] = keyVal[1]
	}

	return parametersMap
}

func (ctx *TestContext) ThereIsAGroupWith(parameters string) error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		group := getParametersMap(parameters)

		err = store.Groups().InsertMap(map[string]interface{}{
			"id":   group["id"],
			"name": "Group " + group["id"].(string),
		})
		if err != nil {
			return err
		}

		return store.GroupAncestors().InsertMap(map[string]interface{}{
			"ancestor_group_id": group["id"],
			"child_group_id":    group["id"],
		})
	})
}

func (ctx *TestContext) ICanWatchParticipantWithID(participantID int64) error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		watchedGroupID := rand.Int63()
		err = store.Groups().InsertMap(map[string]interface{}{
			"id":   watchedGroupID,
			"name": strconv.FormatInt(ctx.userID, 10) + " watch " + strconv.FormatInt(participantID, 10),
		})
		if err != nil {
			return err
		}

		err = store.GroupGroups().InsertMap(map[string]interface{}{
			"parent_group_id": watchedGroupID,
			"child_group_id":  participantID,
		})
		if err != nil {
			return err
		}

		err = store.GroupAncestors().InsertMaps([]map[string]interface{}{
			{"ancestor_group_id": watchedGroupID, "child_group_id": participantID},
			{"ancestor_group_id": watchedGroupID, "child_group_id": watchedGroupID},
		})
		if err != nil {
			return err
		}

		return store.GroupManagers().InsertMap(map[string]interface{}{
			"manager_id":        ctx.userID,
			"group_id":          watchedGroupID,
			"can_watch_members": true,
		})
	})
}

func (ctx *TestContext) IAmInTheGroupWithID(groupID int64) error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		err = ctx.ThereIsAGroupWith("id=" + strconv.FormatInt(groupID, 10))
		if err != nil {
			return err
		}

		err = store.GroupGroups().InsertMap(map[string]interface{}{
			"parent_group_id": groupID,
			"child_group_id":  ctx.userID,
		})
		if err != nil {
			return err
		}

		return store.GroupAncestors().InsertMap(map[string]interface{}{
			"ancestor_group_id": groupID,
			"child_group_id":    ctx.userID,
		})
	})
}

func (ctx *TestContext) ICanOnItemWithID(watchType, watchValue string, itemID int64) error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		return database.NewDataStoreWithTable(store.DB, "permissions_generated").InsertMap(map[string]interface{}{
			"group_id":                        ctx.userID,
			"item_id":                         itemID,
			"can_" + watchType + "_generated": watchValue,
		})
	})
}

func (ctx *TestContext) ICanViewOnItemWithID(watchValue string, itemID int64) error {
	return ctx.ICanOnItemWithID("view", watchValue, itemID)
}

func (ctx *TestContext) ICanWatchOnItemWithID(watchValue string, itemID int64) error {
	return ctx.ICanOnItemWithID("watch", watchValue, itemID)
}

func (ctx *TestContext) IHaveValidatedItemWithID(itemID int64) error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		attemptID := rand.Int63()
		err = store.Attempts().InsertMap(map[string]interface{}{
			"id":             attemptID,
			"participant_id": ctx.userID,
		})
		if err != nil {
			return err
		}

		return store.Results().InsertMap(map[string]interface{}{
			"attempt_id":     attemptID,
			"participant_id": ctx.userID,
			"item_id":        itemID,
			"validated_at":   time.Now(),
		})
	})
}

func (ctx *TestContext) ThereIsAThreadWith(parameters string) error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		thread := map[string]interface{}{}
		arrayParameters := strings.Split(parameters, ",")
		for _, paramKeyValue := range arrayParameters {
			keyVal := strings.Split(paramKeyValue, "=")
			thread[keyVal[0]] = keyVal[1]
		}

		// add item
		err = store.Items().InsertMap(map[string]interface{}{
			"id":                   thread["itemID"],
			"default_language_tag": "en",
			"type":                 "Task",
		})
		if err != nil {
			return err
		}

		// add helper_group_id
		if _, ok := thread["helper_group_id"]; !ok {
			helperGroupID := rand.Int63()
			err = store.Groups().InsertMap(map[string]interface{}{
				"id":   helperGroupID,
				"name": "helper_group_for_" + thread["item_id"].(string) + "," + thread["participant_id"].(string),
			})
			if err != nil {
				return err
			}

			err = store.GroupAncestors().InsertMap(map[string]interface{}{
				"ancestor_group_id": helperGroupID,
				"child_group_id":    helperGroupID,
			})
			if err != nil {
				return err
			}

			thread["helper_group_id"] = helperGroupID
		}

		ctx.thread = thread

		return store.Threads().InsertMap(thread)
	})
}
func (ctx *TestContext) ThereIsNoThreadWith(parameters string) error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		thread := map[string]interface{}{}
		arrayParameters := strings.Split(parameters, ",")
		for _, paramKeyValue := range arrayParameters {
			keyVal := strings.Split(paramKeyValue, "=")
			thread[keyVal[0]] = keyVal[1]
		}

		// add item
		return store.Items().InsertMap(map[string]interface{}{
			"id":                   thread["item_id"],
			"default_language_tag": "en",
			"type":                 "Task",
		})
	})
}

func (ctx *TestContext) IAmPartOfTheHelperGroupOfTheThread() error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		err = store.GroupGroups().InsertMap(map[string]interface{}{
			"parent_group_id": ctx.thread["helper_group_id"],
			"child_group_id":  ctx.userID,
		})
		if err != nil {
			return err
		}

		return store.GroupAncestors().InsertMap(map[string]interface{}{
			"ancestor_group_id": ctx.thread["helper_group_id"],
			"child_group_id":    ctx.userID,
		})
	})
}

func (ctx *TestContext) ICanRequestHelpToTheGroupWithIDOnTheItemWithID(groupID, itemID int64) error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		return store.PermissionsGranted().InsertMap(map[string]interface{}{
			"group_id":            ctx.userID,
			"source_group_id":     ctx.userID,
			"item_id":             itemID,
			"can_request_help_to": groupID,
		})
	})
}
