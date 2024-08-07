//go:build !prod

package testhelpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
	"github.com/France-ioi/AlgoreaBackend/v2/app/utils"
)

const (
	referencePrefix = '@'
	strTrue         = "true"
)

var (
	itemPermissionKeys  = []string{"can_view", "can_grant_view", "can_watch", "can_edit", "is_owner", "can_request_help_to"}
	itemPropagationKeys = []string{"grant_view_propagation", "watch_propagation", "edit_propagation", "request_help_propagation"}
)

// ctx.getParameterMap parses parameters in format key1=val1,key2=val2,... into a map.
func (ctx *TestContext) getParameterMap(parameters string) map[string]string {
	parameterMap := make(map[string]string)
	arrayParameters := strings.Split(parameters, ",")
	for _, paramKeyValue := range arrayParameters {
		keyVal := strings.Split(paramKeyValue, "=")

		parameterMap[keyVal[0]] = keyVal[1]
	}

	return parameterMap
}

// getParameterString converts parameters into a string with format key1=val1,key2=val2,...
func getParameterString(parameters map[string]string) string {
	var str string
	for key, value := range parameters {
		if str != "" {
			str += ","
		}
		str += key + "=" + value
	}

	return str
}

// referenceToName returns the name of a reference.
func referenceToName(reference string) string {
	if reference == "" {
		return ""
	}
	if reference[0] == referencePrefix {
		return reference[1:]
	}

	return reference
}

// getRowMap convert a PickleTable's row into a map where the keys are the column headers.
func (ctx *TestContext) getRowMap(rowIndex int, table *godog.Table) map[string]string {
	rowHeader := table.Rows[0]
	sourceRow := table.Rows[rowIndex]

	rowMap := map[string]string{}
	for i := 0; i < len(rowHeader.Cells); i++ {
		value := sourceRow.Cells[i].Value
		if value == "" {
			continue
		}

		rowMap[rowHeader.Cells[i].Value] = value
	}

	return rowMap
}

// populateDatabase populate the database with all the initialized data.
func (ctx *TestContext) populateDatabase() error {
	// We cannot run this for older tests because we're computing the tables permissions_generated and item_ancestors.
	// Older tests define those tables manually with inconsistencies, and then check that the content of those tables is
	// still in the same inconsistent state.
	// If we want this to be run everywhere, we would have to fix those tests first.
	// We would then just have to remove the ctx.needPopulateDatabase boolean completely.
	if !ctx.needPopulateDatabase {
		return nil
	}

	db, err := database.Open(ctx.db)
	if err != nil {
		return err
	}

	// add all the defined table rows in the database.
	err = database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		err = ctx.addUsersIntoAllUsersGroup()
		if err != nil {
			return err
		}

		for tableName, tableRows := range ctx.dbTables {
			for _, tableRow := range tableRows {
				err = database.NewDataStoreWithTable(store.DB, tableName).InsertOrUpdateMap(tableRow, nil)
				if err != nil {
					return fmt.Errorf("populateDatabase %s %+v: %v", tableName, tableRow, err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = ctx.DBItemsAncestorsAndPermissionsAreComputed()
	if err != nil {
		return err
	}

	return ctx.DBGroupsAncestorsAreComputed()
}

func (ctx *TestContext) isInDatabase(tableName, key string) bool {
	if ctx.dbTables[tableName] == nil {
		return false
	}

	_, ok := ctx.dbTables[tableName][key]
	return ok
}

func mergeFields(oldValues, newValues map[string]interface{}) map[string]interface{} {
	merged := oldValues
	for key, newValue := range newValues {
		merged[key] = newValue
	}

	return merged
}

func (ctx *TestContext) addToDatabase(tableName, key string, row map[string]interface{}) {
	ctx.needPopulateDatabase = true

	if ctx.dbTables[tableName] == nil {
		ctx.dbTables[tableName] = make(map[string]map[string]interface{})
	}

	ctx.dbTables[tableName][key] = row
}

// addPersonalInfoViewApprovedFor adds a permission generated in the database.
func (ctx *TestContext) addPersonalInfoViewApprovedFor(childGroup, parentGroup string) {
	parentGroupID := ctx.getIDOfReference(parentGroup)
	childGroupID := ctx.getIDOfReference(childGroup)

	groupGroupTable := "groups_groups"
	key := ctx.getGroupGroupKey(parentGroupID, childGroupID)
	if !ctx.isInDatabase(groupGroupTable, key) {
		ctx.addGroupGroup(parentGroup, childGroup)
	}

	ctx.dbTables[groupGroupTable][key]["personal_info_view_approved_at"] = time.Now()
}

// getGroupGroupKey gets a group group unique key for the groupgroup's map.
func (ctx *TestContext) getGroupGroupKey(parentGroupID, childGroupID int64) string {
	return strconv.FormatInt(parentGroupID, 10) + "," + strconv.FormatInt(childGroupID, 10)
}

// addGroupGroup adds a group-group in the database.
func (ctx *TestContext) addGroupGroup(parentGroup, childGroup string) {
	parentGroupID := ctx.getIDOfReference(parentGroup)
	childGroupID := ctx.getIDOfReference(childGroup)

	ctx.addToDatabase("groups_groups", ctx.getGroupGroupKey(parentGroupID, childGroupID), map[string]interface{}{
		"parent_group_id": parentGroupID,
		"child_group_id":  childGroupID,
	})
}

// addGroupManager adds a group manager in the database.
func (ctx *TestContext) addGroupManager(manager, group, canWatchMembers, canGrantGroupAccess, canManage string) {
	managerID := ctx.getIDOfReference(manager)
	groupID := ctx.getIDOfReference(group)

	ctx.addToDatabase(
		"group_managers",
		strconv.FormatInt(managerID, 10)+","+strconv.FormatInt(groupID, 10),
		map[string]interface{}{
			"manager_id":             managerID,
			"group_id":               groupID,
			"can_watch_members":      canWatchMembers,
			"can_grant_group_access": canGrantGroupAccess,
			"can_manage":             canManage,
		},
	)
}

// addPermissionsGranted adds a permission granted in the database.
func (ctx *TestContext) addPermissionGranted(group, item, sourceGroup, origin, permission, permissionValue string) {
	groupID := ctx.getIDOfReference(group)
	sourceGroupID := ctx.getIDOfReference(sourceGroup)
	itemID := ctx.getIDOfReference(item)

	permissionsGrantedTable := "permissions_granted"
	key := strconv.FormatInt(groupID, 10) + "," +
		strconv.FormatInt(itemID, 10) + "," +
		strconv.FormatInt(sourceGroupID, 10) + "," +
		origin

	if !ctx.isInDatabase(permissionsGrantedTable, key) {
		ctx.addToDatabase(permissionsGrantedTable, key, map[string]interface{}{
			"group_id":        groupID,
			"source_group_id": sourceGroupID,
			"item_id":         itemID,
			"origin":          origin,
		})
	}

	if permission == "can_request_help_to" {
		canRequestHelpToGroupID := ctx.getIDOfReference(permissionValue)

		err := ctx.ThereIsAGroup(permissionValue)
		if err != nil {
			return
		}

		permissionValue = strconv.FormatInt(canRequestHelpToGroupID, 10)
	}

	if permission == "is_owner" {
		boolValue, err := strconv.ParseBool(permissionValue)
		if err != nil {
			panic(fmt.Sprintf("%v cannot be parsed as a boolean", boolValue))
		}

		ctx.dbTables[permissionsGrantedTable][key][permission] = boolValue
	} else {
		ctx.dbTables[permissionsGrantedTable][key][permission] = permissionValue
	}
}

// addAttempt adds an attempt to the database.
func (ctx *TestContext) addAttempt(item, participant string) {
	itemID := ctx.getIDOfReference(item)
	participantID := ctx.getIDOfReference(participant)

	ctx.addToDatabase(
		`attempts`,
		strconv.FormatInt(itemID, 10)+","+strconv.FormatInt(participantID, 10),
		map[string]interface{}{
			"id":             itemID,
			"participant_id": participantID,
		},
	)
}

// addValidatedResult adds a validated result to the database.
func (ctx *TestContext) addValidatedResult(attemptID, participant, item string, validatedAt time.Time) {
	participantID := ctx.getIDOfReference(participant)
	itemID := ctx.getIDOfReference(item)

	ctx.addToDatabase(
		"results",
		attemptID+","+strconv.FormatInt(participantID, 10)+","+strconv.FormatInt(itemID, 10),
		map[string]interface{}{
			"attempt_id":     attemptID,
			"participant_id": participantID,
			"item_id":        itemID,
			"validated_at":   validatedAt,
		},
	)
}

func (ctx *TestContext) getItemItemKey(parentItemID, childItemID int64) string {
	return strconv.FormatInt(parentItemID, 10) + "," + strconv.FormatInt(childItemID, 10)
}

// addItemItem adds an item-item in the database.
func (ctx *TestContext) addItemItem(parentItem, childItem string) {
	parentItemID := ctx.getIDOfReference(parentItem)
	childItemID := ctx.getIDOfReference(childItem)

	ctx.addToDatabase(
		"items_items",
		ctx.getItemItemKey(parentItemID, childItemID),
		map[string]interface{}{
			"parent_item_id": parentItemID,
			"child_item_id":  childItemID,
			"child_order":    rand.Int31n(1000),
		},
	)
}

func (ctx *TestContext) addItemItemPropagation(parent, child, propagation, propagationValue string) {
	key := ctx.getItemItemKey(ctx.getIDOfReference(parent), ctx.getIDOfReference(child))

	ctx.dbTables["items_items"][key][propagation] = propagationValue
}

// addItem adds an item in the database.
func (ctx *TestContext) addItem(fields map[string]string) {
	dbFields := make(map[string]interface{})
	for key, value := range fields {
		if key == "item" {
			key = "id"
		}

		switch {
		case strings.HasSuffix(key, "id"):
			dbFields[key] = ctx.getIDOfReference(value)
		case value[0] == referencePrefix:
			dbFields[key] = value[1:]
		default:
			dbFields[key] = value
		}
	}

	itemKey := strconv.FormatInt(dbFields["id"].(int64), 10)

	if oldFields, ok := ctx.dbTables["items"][itemKey]; ok {
		dbFields = mergeFields(oldFields, dbFields)
	}

	if _, ok := dbFields["type"]; !ok {
		dbFields["type"] = "Task"
	}
	if _, ok := dbFields["default_language_tag"]; !ok {
		dbFields["default_language_tag"] = "en"
	}
	if _, ok := dbFields["text_id"]; !ok && fields["item"][0] == referencePrefix {
		dbFields["text_id"] = fields["item"][1:]
	}

	ctx.addToDatabase("items", itemKey, dbFields)
}

// getThreadKey gets a thread unique key for the thread's map.
func (ctx *TestContext) getThreadKey(itemID, participantID int64) string {
	return strconv.FormatInt(itemID, 10) + "," + strconv.FormatInt(participantID, 10)
}

// addThread adds a thread to the database.
func (ctx *TestContext) addThread(item, participant, helperGroup, status, messageCount, latestUpdateAt string) {
	itemID := ctx.getIDOfReference(item)
	participantID := ctx.getIDOfReference(participant)
	helperGroupID := ctx.getIDOfReference(helperGroup)

	latestUpdateAtDate, err := time.Parse(utils.DateTimeFormat, latestUpdateAt)
	if err != nil {
		panic(err)
	}

	ctx.addToDatabase("threads", ctx.getThreadKey(itemID, participantID), map[string]interface{}{
		"item_id":          itemID,
		"participant_id":   participantID,
		"helper_group_id":  helperGroupID,
		"status":           status,
		"message_count":    messageCount,
		"latest_update_at": latestUpdateAtDate,
	})
}

// UserIsAManagerOfTheGroupWith sets the current user as the manager of a group.
func (ctx *TestContext) UserIsAManagerOfTheGroupWith(parameters string) error {
	group := ctx.getParameterMap(parameters)

	err := ctx.ThereIsAGroup(group["id"])
	if err != nil {
		return err
	}

	canWatchMembers := "0"
	canGrantGroupAccess := "0"
	canManage := "none"

	if group["can_watch_members"] == strTrue {
		canWatchMembers = "1"
	}
	if group["can_grant_group_access"] == strTrue {
		canGrantGroupAccess = "1"
	}
	if _, ok := group["can_manage"]; ok {
		canManage = group["can_manage"]
	}

	ctx.addGroupManager(group["user_id"], group["id"], canWatchMembers, canGrantGroupAccess, canManage)

	return nil
}

// IAmAManagerOfTheGroupWithID sets the user as a manager of a group with an id.
func (ctx *TestContext) IAmAManagerOfTheGroupWithID(group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                group,
		"user_id":           ctx.user,
		"can_watch_members": "false",
	}))
}

// IAmAManagerOfTheGroup sets the user as a manager of a group with an id.
func (ctx *TestContext) IAmAManagerOfTheGroup(group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                group,
		"user_id":           ctx.user,
		"name":              group,
		"can_watch_members": "false",
	}))
}

// IAmAManagerOfTheGroupAndCanWatchItsMembers sets the user as a manager of a group with can_watch permission.
func (ctx *TestContext) IAmAManagerOfTheGroupAndCanWatchItsMembers(group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                group,
		"user_id":           ctx.user,
		"name":              group,
		"can_watch_members": strTrue,
	}))
}

// UserIsAManagerOfTheGroupAndCanWatchItsMembers sets the user as a manager of the group with can_watch permission.
func (ctx *TestContext) UserIsAManagerOfTheGroupAndCanWatchItsMembers(user, group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                group,
		"user_id":           user,
		"name":              group,
		"can_watch_members": strTrue,
	}))
}

// UserIsAManagerOfTheGroupAndCanGrantGroupAccess sets the user as a manager of the group with can_grant_group_access permission.
func (ctx *TestContext) UserIsAManagerOfTheGroupAndCanGrantGroupAccess(user, group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                     group,
		"user_id":                user,
		"name":                   group,
		"can_grant_group_access": strTrue,
	}))
}

// UserIsAManagerOfTheGroupAndCanManageMembershipsAndGroup adds the user as a manager of the group
// with the can_manage=memberships_and_groups permission.
func (ctx *TestContext) UserIsAManagerOfTheGroupAndCanManageMembershipsAndGroup(user, group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":         group,
		"user_id":    user,
		"name":       group,
		"can_manage": "memberships_and_group",
	}))
}

// ICanWatchGroupWithID adds the permission for the user to watch a group.
func (ctx *TestContext) ICanWatchGroupWithID(group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                group,
		"user_id":           ctx.user,
		"can_watch_members": strTrue,
	}))
}

// ThereAreTheFollowingItems defines items.
func (ctx *TestContext) ThereAreTheFollowingItems(items *godog.Table) error {
	for i := 1; i < len(items.Rows); i++ {
		item := ctx.getRowMap(i, items)

		ctx.addItem(map[string]string{
			"item": item["item"],
			"type": item["type"],
		})

		if _, ok := item["parent"]; ok {
			parents := strings.Split(item["parent"], ",")

			for _, parent := range parents {
				ctx.addItemItem(parent, item["item"])
			}
		}
	}

	return nil
}

// ThereAreTheFollowingTasks defines item tasks.
func (ctx *TestContext) ThereAreTheFollowingTasks(tasks *godog.Table) error {
	for i := 1; i < len(tasks.Rows); i++ {
		task := ctx.getRowMap(i, tasks)

		ctx.addItem(map[string]string{
			"item": task["item"],
			"type": "Task",
		})
	}

	return nil
}

// ThereAreTheFollowingItemPermissions defines item permissions.
func (ctx *TestContext) ThereAreTheFollowingItemPermissions(itemPermissions *godog.Table) error {
	for i := 1; i < len(itemPermissions.Rows); i++ {
		itemPermission := ctx.getRowMap(i, itemPermissions)

		err := ctx.applyUserPermissionsOnItem(itemPermission)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx *TestContext) applyUserPermissionsOnItem(itemPermission map[string]string) error {
	sourceGroup := itemPermission["group"]
	if definedSourceGroup, ok := itemPermission["source_group"]; ok {
		sourceGroup = definedSourceGroup
	}

	origin := "group_membership"
	if definedOrigin, ok := itemPermission["origin"]; ok {
		origin = definedOrigin
	}

	for _, permissionKey := range itemPermissionKeys {
		if permissionValue, ok := itemPermission[permissionKey]; ok {
			err := ctx.UserSetPermissionExtendedOnItemWithID(
				permissionKey,
				permissionValue,
				itemPermission["group"],
				itemPermission["item"],
				sourceGroup,
				origin,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ThereAreTheFollowingItemRelations defines item relations, in items_items table.
func (ctx *TestContext) ThereAreTheFollowingItemRelations(itemPermissions *godog.Table) error {
	for i := 1; i < len(itemPermissions.Rows); i++ {
		itemRelation := ctx.getRowMap(i, itemPermissions)

		err := ctx.applyItemRelation(itemRelation)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx *TestContext) applyItemRelation(itemRelation map[string]string) error {
	ctx.addItemItem(itemRelation["parent"], itemRelation["item"])

	for _, propagationKey := range itemPropagationKeys {
		if propagationValue, ok := itemRelation[propagationKey]; ok {
			boolValue, err := strconv.ParseBool(propagationValue)
			if err != nil {
				panic(fmt.Sprintf("applyItemRelation: %v cannot be parsed as a boolean", boolValue))
			}

			if boolValue {
				propagationValue = "1"
			} else {
				propagationValue = "0"
			}

			ctx.addItemItemPropagation(
				itemRelation["parent"],
				itemRelation["item"],
				propagationKey,
				propagationValue,
			)
		}
	}

	return nil
}

// ICanWatchGroup adds the permission for the user to watch a group.
func (ctx *TestContext) ICanWatchGroup(groupName string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                groupName,
		"user_id":           ctx.user,
		"name":              groupName,
		"can_watch_members": strTrue,
	}))
}

// IsAMemberOfTheGroup Puts a group in a group.
func (ctx *TestContext) IsAMemberOfTheGroup(childGroupName, parentGroupName string) {
	ctx.addGroupGroup(parentGroupName, childGroupName)
}

// ItemRelationSetPropagation adds a propagation on an item relation.
func (ctx *TestContext) ItemRelationSetPropagation(propagation, value, parent, item string) error {
	ctx.addItemItemPropagation(parent, item, propagation, value)

	return nil
}

// UserSetPermissionExtendedOnItemWithID gives a user a permission on an item with a specific source_group and origin.
func (ctx *TestContext) UserSetPermissionExtendedOnItemWithID(permission, value, user, item, sourceGroup, origin string) error {
	ctx.addPermissionGranted(user, item, sourceGroup, origin, permission, value)

	return nil
}

// UserSetPermissionOnItemWithID gives a user a permission on an item.
func (ctx *TestContext) UserSetPermissionOnItemWithID(permission, value, user, item string) error {
	ctx.addPermissionGranted(user, item, user, "group_membership", permission, value)

	return nil
}

// ICanOnItemWithID gives the user a permission on an item.
func (ctx *TestContext) ICanOnItemWithID(watchType, watchValue, item string) error {
	return ctx.UserSetPermissionOnItemWithID(watchType, watchValue, ctx.user, item)
}

// UserCanViewOnItemWithID gives a user a can_view permission on an item.
func (ctx *TestContext) UserCanViewOnItemWithID(viewValue, user, item string) error {
	return ctx.UserSetPermissionOnItemWithID("can_view", viewValue, user, item)
}

// ICanViewOnItemWithID gives the user a "view" permission on an item.
func (ctx *TestContext) ICanViewOnItemWithID(viewValue, item string) error {
	return ctx.UserSetPermissionOnItemWithID("can_view", viewValue, ctx.user, item)
}

// UserCanGrantViewOnItemWithID gives a user a can_grant_view permission on an item.
func (ctx *TestContext) UserCanGrantViewOnItemWithID(viewValue, user, item string) error {
	return ctx.UserSetPermissionOnItemWithID("can_grant_view", viewValue, user, item)
}

// UserCanWatchOnItemWithID gives a user a "watch" permission on an item.
func (ctx *TestContext) UserCanWatchOnItemWithID(watchValue, user, item string) error {
	return ctx.UserSetPermissionOnItemWithID("can_watch", watchValue, user, item)
}

// ICanWatchOnItemWithID gives the user a "watch" permission on an item.
func (ctx *TestContext) ICanWatchOnItemWithID(watchValue, item string) error {
	return ctx.UserSetPermissionOnItemWithID("can_watch", watchValue, ctx.user, item)
}

// UserIsOwnerOfItemWithID sets the is_owner permission.
func (ctx *TestContext) UserIsOwnerOfItemWithID(isOwner, user, item string) error {
	return ctx.UserSetPermissionOnItemWithID("is_owner", isOwner, user, item)
}

// UserCanRequestHelpToOnItemWithID sets the can_request_help_to permission.
func (ctx *TestContext) UserCanRequestHelpToOnItemWithID(canRequestHelpTo, user, item string) error {
	return ctx.UserSetPermissionOnItemWithID("can_request_help_to", canRequestHelpTo, user, item)
}

func (ctx *TestContext) userHasAValidatedItemWithID(user, item string) error {
	attemptID := rand.Int63()

	ctx.addAttempt(item, user)
	ctx.addValidatedResult(
		strconv.FormatInt(attemptID, 10),
		user,
		item,
		time.Now(),
	)

	return nil
}

// ThereAreTheFollowingResults creates validated results described in the given Godog table.
func (ctx *TestContext) ThereAreTheFollowingResults(results *godog.Table) error {
	for i := 1; i < len(results.Rows); i++ {
		result := ctx.getRowMap(i, results)

		ctx.addItem(map[string]string{
			"item": result["item"],
		})

		err := ctx.userHasAValidatedItemWithID(result["participant"], result["item"])
		if err != nil {
			return err
		}
	}

	return nil
}

// IHaveValidatedItemWithID states that user has validated an item.
func (ctx *TestContext) IHaveValidatedItemWithID(item string) error {
	return ctx.userHasAValidatedItemWithID(ctx.user, item)
}

// ThereAreTheFollowingThreads create threads.
func (ctx *TestContext) ThereAreTheFollowingThreads(threads *godog.Table) error {
	for i := 1; i < len(threads.Rows); i++ {
		thread := ctx.getRowMap(i, threads)
		threadParameters := make(map[string]string)

		threadParameters["participant_id"] = thread["participant"]

		if thread["item"] != "" {
			threadParameters["item_id"] = thread["item"]
		}

		if thread["helper_group"] != "" {
			threadParameters["helper_group_id"] = thread["helper_group"]
		}

		if thread["status"] != "" {
			threadParameters["status"] = thread["status"]
		}

		if thread["latest_update_at"] != "" {
			threadParameters["latest_update_at"] = thread["latest_update_at"]
		}

		if thread["message_count"] != "" {
			threadParameters["message_count"] = thread["message_count"]
		}

		if thread["visible_by_participant"] == "1" {
			err := ctx.UserCanViewOnItemWithID("content", thread["participant"], thread["item"])
			if err != nil {
				return err
			}
		}

		err := ctx.ThereIsAThreadWith(getParameterString(threadParameters))
		if err != nil {
			return err
		}
	}

	return nil
}

// ThereIsAThreadWith creates a thread.
func (ctx *TestContext) ThereIsAThreadWith(parameters string) error {
	thread := ctx.getParameterMap(parameters)

	// add item
	if _, ok := thread["item_id"]; !ok {
		thread["item_id"] = strconv.FormatInt(rand.Int63(), 10)
	}

	ctx.addItem(map[string]string{
		"item": thread["item_id"],
	})

	// add helper_group_id
	if _, ok := thread["helper_group_id"]; !ok {
		helperGroupID := rand.Int63()

		err := ctx.ThereIsAGroup(strconv.FormatInt(helperGroupID, 10))
		if err != nil {
			return err
		}
		ctx.setGroupFieldInDatabase(
			ctx.getGroupPrimaryKey(helperGroupID),
			"name",
			"helper_group_for_"+thread["item_id"]+"-"+thread["participant_id"],
		)

		thread["helper_group_id"] = strconv.FormatInt(helperGroupID, 10)
	}

	// add status
	if _, ok := thread["status"]; !ok {
		thread["status"] = "waiting_for_trainer"
	}

	// add message count
	if _, ok := thread["message_count"]; !ok {
		thread["message_count"] = "0"
	}

	// add latest update at
	if _, ok := thread["latest_update_at"]; !ok {
		thread["latest_update_at"] = time.Now().Format(utils.DateTimeFormat)
	}

	ctx.currentThreadKey = ctx.getThreadKey(
		ctx.getIDOfReference(thread["item_id"]),
		ctx.getIDOfReference(thread["participant_id"]),
	)

	ctx.addThread(
		thread["item_id"],
		thread["participant_id"],
		thread["helper_group_id"],
		thread["status"],
		thread["message_count"],
		thread["latest_update_at"],
	)

	return nil
}

// ThereIsNoThreadWith states that a given thread doesn't exist.
func (ctx *TestContext) ThereIsNoThreadWith(parameters string) error {
	thread := ctx.getParameterMap(parameters)

	ctx.addItem(map[string]string{
		"item": thread["item_id"],
	})

	return nil
}

// IAmPartOfTheHelperGroupOfTheThread states that user is a member of the helper group, of a given thread.
func (ctx *TestContext) IAmPartOfTheHelperGroupOfTheThread() error {
	threadHelperGroupID := ctx.dbTables["threads"][ctx.currentThreadKey]["helper_group_id"].(int64)

	ctx.IsAMemberOfTheGroup(
		ctx.user,
		strconv.FormatInt(threadHelperGroupID, 10),
	)

	return nil
}

// ICanRequestHelpToTheGroupWithIDOnTheItemWithID gives the user the permission to request help from a given group
// to a given item.
func (ctx *TestContext) ICanRequestHelpToTheGroupWithIDOnTheItemWithID(group, item string) error {
	return ctx.UserCanRequestHelpToOnItemWithID(group, ctx.user, item)
}
