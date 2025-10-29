//go:build !prod && !unit

package testhelpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
)

const (
	referencePrefix = '@'
	idString        = "id"
)

// ctx.getParameterMap parses parameters in format key1=val1,key2=val2,... into a map.
func (ctx *TestContext) getParameterMap(parameters string) map[string]string {
	preprocessed := ctx.preprocessString(parameters)

	parameterMap := make(map[string]string)
	arrayParameters := strings.Split(preprocessed, ",")
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

	rowMap := make(map[string]string, len(rowHeader.Cells))
	for cellIndex := 0; cellIndex < len(rowHeader.Cells); cellIndex++ {
		value := sourceRow.Cells[cellIndex].Value
		if value == "" {
			continue
		}

		rowMap[rowHeader.Cells[cellIndex].Value] = value
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

	err := ctx.DBItemsAncestorsAreComputed()
	if err != nil {
		return err
	}

	err = ctx.DBGroupsAncestorsAreComputed()
	if err != nil {
		return err
	}

	err = ctx.DBGeneratedPermissionsAreComputed()
	if err != nil {
		return err
	}

	return ctx.DBResultsAreComputed()
}

func (ctx *TestContext) isInDatabase(tableName string, primaryKey map[string]string) bool {
	return ctx.getDBTableRowIndexForPrimaryKey(tableName, primaryKey) != -1
}

// addPersonalInfoViewApprovedFor adds a permission generated in the database.
func (ctx *TestContext) addPersonalInfoViewApprovedFor(childGroup, parentGroup string) {
	parentGroupID := ctx.getIDOrIDByReference(parentGroup)
	childGroupID := ctx.getIDOrIDByReference(childGroup)

	const groupGroupTable = "groups_groups"
	key := ctx.getGroupGroupKey(parentGroupID, childGroupID)
	if !ctx.isInDatabase(groupGroupTable, key) {
		ctx.addGroupGroup(parentGroup, childGroup)
	}

	ctx.setDBTableRowColumnValue(groupGroupTable, key, "personal_info_view_approved_at", time.Now().UTC().Format(time.DateTime))
}

// getGroupGroupKey constructs a key for a searching a group-group pair in the groups_groups table.
func (ctx *TestContext) getGroupGroupKey(parentGroupID, childGroupID int64) map[string]string {
	return map[string]string{
		"parent_group_id": strconv.FormatInt(parentGroupID, 10),
		"child_group_id":  strconv.FormatInt(childGroupID, 10),
	}
}

// addGroupGroup adds a group-group in the database.
func (ctx *TestContext) addGroupGroup(parentGroup, childGroup string) {
	parentGroupID := ctx.getIDOrIDByReference(parentGroup)
	childGroupID := ctx.getIDOrIDByReference(childGroup)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("groups_groups",
		constructGodogTableFromData([]stringKeyValuePair{
			{"parent_group_id", strconv.FormatInt(parentGroupID, 10)},
			{"child_group_id", strconv.FormatInt(childGroupID, 10)},
		}))
	if err != nil {
		panic(err)
	}
}

// addGroupManager adds a group manager in the database.
func (ctx *TestContext) addGroupManager(manager, group, canWatchMembers, canGrantGroupAccess, canManage string) error {
	err := ctx.ThereIsAGroup(group)
	if err != nil {
		return err
	}

	managerID := ctx.getIDOrIDByReference(manager)
	groupID := ctx.getIDOrIDByReference(group)

	ctx.needPopulateDatabase = true
	err = ctx.DBHasTable("group_managers",
		constructGodogTableFromData([]stringKeyValuePair{
			{"manager_id", strconv.FormatInt(managerID, 10)},
			{"group_id", strconv.FormatInt(groupID, 10)},
			{"can_watch_members", canWatchMembers},
			{"can_grant_group_access", canGrantGroupAccess},
			{"can_manage", canManage},
		}))
	if err != nil {
		panic(err)
	}
	return nil
}

// setGrantedPermission sets a granted permission in the database.
func (ctx *TestContext) setGrantedPermission(group, item, sourceGroup, origin, permission, permissionValue string) {
	groupIDString := strconv.FormatInt(ctx.getIDOrIDByReference(group), 10)
	sourceGroupIDString := strconv.FormatInt(ctx.getIDOrIDByReference(sourceGroup), 10)
	itemIDString := strconv.FormatInt(ctx.getIDOrIDByReference(item), 10)

	const permissionsGrantedTable = "permissions_granted"
	primaryKey := map[string]string{
		"group_id": groupIDString, "item_id": itemIDString, "source_group_id": sourceGroupIDString, "origin": origin,
	}

	if !ctx.isInDatabase(permissionsGrantedTable, primaryKey) {
		ctx.needPopulateDatabase = true
		err := ctx.DBHasTable(permissionsGrantedTable,
			constructGodogTableFromData([]stringKeyValuePair{
				{"group_id", groupIDString},
				{"source_group_id", sourceGroupIDString},
				{"item_id", itemIDString},
				{"origin", origin},
			}))
		if err != nil {
			panic(err)
		}
	}

	if permission == "can_request_help_to" {
		canRequestHelpToGroupID := ctx.getIDOrIDByReference(permissionValue)

		err := ctx.ThereIsAGroup(permissionValue)
		if err != nil {
			return
		}

		permissionValue = strconv.FormatInt(canRequestHelpToGroupID, 10)
	}

	ctx.setDBTableRowColumnValue(permissionsGrantedTable, primaryKey, permission, permissionValue)
}

// addAttempt adds an attempt to the database.
func (ctx *TestContext) addAttempt(item, participant string) {
	itemID := ctx.getIDOrIDByReference(item)
	participantID := ctx.getIDOrIDByReference(participant)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("attempts",
		constructGodogTableFromData([]stringKeyValuePair{
			{idString, strconv.FormatInt(itemID, 10)},
			{"participant_id", strconv.FormatInt(participantID, 10)},
		}))
	if err != nil {
		panic(err)
	}
}

// addValidatedResult adds a validated result to the database.
func (ctx *TestContext) addValidatedResult(attemptID, participant, item string, validatedAt time.Time) {
	participantID := ctx.getIDOrIDByReference(participant)
	itemID := ctx.getIDOrIDByReference(item)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("results",
		constructGodogTableFromData([]stringKeyValuePair{
			{"attempt_id", attemptID},
			{"participant_id", strconv.FormatInt(participantID, 10)},
			{"item_id", strconv.FormatInt(itemID, 10)},
			{"validated_at", validatedAt.UTC().Format(time.DateTime)},
		}))
	if err != nil {
		panic(err)
	}
}

func (ctx *TestContext) getItemItemPrimaryKey(parentItemID, childItemID int64) map[string]string {
	return map[string]string{
		"parent_item_id": strconv.FormatInt(parentItemID, 10),
		"child_item_id":  strconv.FormatInt(childItemID, 10),
	}
}

// addItemItem adds an item-item in the database.
func (ctx *TestContext) addItemItem(parentItem, childItem string) {
	parentItemID := ctx.getIDOrIDByReference(parentItem)
	childItemID := ctx.getIDOrIDByReference(childItem)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("items_items",
		constructGodogTableFromData([]stringKeyValuePair{
			{"parent_item_id", strconv.FormatInt(parentItemID, 10)},
			{"child_item_id", strconv.FormatInt(childItemID, 10)},
			{"child_order", strconv.FormatInt((parentItemID+childItemID)%1000, 10)},
		}))
	if err != nil {
		panic(err)
	}
}

func (ctx *TestContext) addItemItemPropagation(parent, child, propagation, propagationValue string) {
	primaryKey := ctx.getItemItemPrimaryKey(ctx.getIDOrIDByReference(parent), ctx.getIDOrIDByReference(child))
	ctx.setDBTableRowColumnValue("items_items", primaryKey, propagation, propagationValue)
}

// addItem adds an item to the database.
func (ctx *TestContext) addItem(fields map[string]string) {
	dbFields, primaryKey, oldRowIndex := ctx.constructDBFieldsForAddItem(fields)

	ctx.needPopulateDatabase = true

	if oldRowIndex != -1 {
		delete(dbFields, idString)
		ctx.setDBTableRowColumnValues("items", primaryKey, dbFields)
		return
	}

	keyValuePairs := make([]stringKeyValuePair, 0, len(dbFields))
	for name, value := range dbFields {
		keyValuePairs = append(keyValuePairs, stringKeyValuePair{name, value})
	}
	err := ctx.DBHasTable("items", constructGodogTableFromData(keyValuePairs))
	if err != nil {
		panic(err)
	}
}

func (ctx *TestContext) constructDBFieldsForAddItem(fields map[string]string) (
	dbFields, primaryKey map[string]string, oldRowIndex int,
) {
	dbFields = make(map[string]string, len(fields))
	for key, value := range fields {
		if key == "item" {
			key = idString
		}

		switch {
		case strings.HasSuffix(key, idString):
			dbFields[key] = strconv.FormatInt(ctx.getIDOrIDByReference(value), 10)
		case value[0] == referencePrefix:
			dbFields[key] = value[1:]
		default:
			dbFields[key] = value
		}
	}

	primaryKey = map[string]string{idString: dbFields[idString]}
	oldRowIndex = ctx.getDBTableRowIndexForPrimaryKey("items", primaryKey)
	ctx.setDefaultValuesInDBFieldsForAddItem(dbFields, fields, oldRowIndex)

	return dbFields, primaryKey, oldRowIndex
}

func (ctx *TestContext) setDefaultValuesInDBFieldsForAddItem(dbFields, fields map[string]string, oldRowIndex int) {
	_, typeIsSet := dbFields["type"]
	_, defaultLanguageTagIsSet := dbFields["default_language_tag"]
	_, textIDIsSet := dbFields["text_id"]

	if oldRowIndex != -1 {
		dbTable := ctx.dbTableData["items"]
		typeIsSet = typeIsSet || isDBTableColumnSetInRow(dbTable, "type", oldRowIndex)
		defaultLanguageTagIsSet = defaultLanguageTagIsSet || isDBTableColumnSetInRow(dbTable, "default_language_tag", oldRowIndex)
		textIDIsSet = typeIsSet || isDBTableColumnSetInRow(dbTable, "text_id", oldRowIndex)
	}

	if !typeIsSet {
		dbFields["type"] = "Task"
	}
	if !defaultLanguageTagIsSet {
		dbFields["default_language_tag"] = "en"
	}
	if !textIDIsSet && fields["item"][0] == referencePrefix {
		dbFields["text_id"] = fields["item"][1:]
	}
}

// getThreadKey gets a thread unique key for the thread's DB table.
func (ctx *TestContext) getThreadKey(itemID, participantID int64) map[string]string {
	return map[string]string{
		"participant_id": strconv.FormatInt(participantID, 10),
		"item_id":        strconv.FormatInt(itemID, 10),
	}
}

// addThread adds a thread to the database.
func (ctx *TestContext) addThread(item, participant, helperGroup, status, messageCount, latestUpdateAt string) {
	itemID := ctx.getIDOrIDByReference(item)
	participantID := ctx.getIDOrIDByReference(participant)
	helperGroupID := ctx.getIDOrIDByReference(helperGroup)

	_, err := time.Parse(time.DateTime, latestUpdateAt)
	if err != nil {
		panic(err)
	}

	err = ctx.DBHasTable("threads",
		constructGodogTableFromData([]stringKeyValuePair{
			{"item_id", strconv.FormatInt(itemID, 10)},
			{"participant_id", strconv.FormatInt(participantID, 10)},
			{"helper_group_id", strconv.FormatInt(helperGroupID, 10)},
			{"status", status},
			{"message_count", messageCount},
			{"latest_update_at", latestUpdateAt},
		}))
	if err != nil {
		panic(err)
	}
}

// IAmAManagerOfTheGroupWithID sets the current user as a manager of a group with an id.
func (ctx *TestContext) IAmAManagerOfTheGroupWithID(group string) error {
	return ctx.addGroupManager(ctx.user, group, "0", "0", "none")
}

// IAmAManagerOfTheGroup sets the current user as a manager of the given group with can_watch_members=false.
func (ctx *TestContext) IAmAManagerOfTheGroup(group string) error {
	return ctx.addGroupManager(ctx.user, group, "0", "0", "none")
}

// GroupIsAManagerOfTheGroupAndCanWatchItsMembers sets the user as a manager of the group with can_watch_members permission.
func (ctx *TestContext) GroupIsAManagerOfTheGroupAndCanWatchItsMembers(managerGroup, group string) error {
	return ctx.addGroupManager(managerGroup, group, "1", "0", "none")
}

// GroupIsAManagerOfTheGroupAndCanGrantGroupAccess sets the group as a manager of the given group with can_grant_group_access permission.
func (ctx *TestContext) GroupIsAManagerOfTheGroupAndCanGrantGroupAccess(managerGroup, group string) error {
	return ctx.addGroupManager(managerGroup, group, "0", "1", "none")
}

// GroupIsAManagerOfTheGroupAndCanManageMembershipsAndGroup sets the group as a manager of the group
// with the can_manage=memberships_and_groups permission.
func (ctx *TestContext) GroupIsAManagerOfTheGroupAndCanManageMembershipsAndGroup(managerGroup, group string) error {
	return ctx.addGroupManager(managerGroup, group, "0", "0", "memberships_and_group")
}

// IAmAManagerOfTheGroupAndCanWatchItsMembers adds the permission for the current user to watch a group.
func (ctx *TestContext) IAmAManagerOfTheGroupAndCanWatchItsMembers(group string) error {
	return ctx.addGroupManager(ctx.user, group, "1", "0", "none")
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

	for _, permissionKey := range []string{
		"can_view", "can_grant_view", "can_watch", "can_edit", "is_owner", "can_request_help_to",
	} {
		if permissionValue, ok := itemPermission[permissionKey]; ok {
			err := ctx.SetGroupPermissionWithSourceGroupAndOriginOnItem(
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

	for _, propagationKey := range []string{
		"grant_view_propagation", "watch_propagation", "edit_propagation", "request_help_propagation",
	} {
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

// GroupIsAMemberOfTheGroup Puts a group into a group.
func (ctx *TestContext) GroupIsAMemberOfTheGroup(childGroupName, parentGroupName string) {
	ctx.addGroupGroup(parentGroupName, childGroupName)
}

// SetGroupPermissionWithSourceGroupAndOriginOnItem gives a group a permission on an item with a specific source_group and origin.
func (ctx *TestContext) SetGroupPermissionWithSourceGroupAndOriginOnItem(permission, value, group, item, sourceGroup, origin string) error {
	ctx.setGrantedPermission(group, item, sourceGroup, origin, permission, value)

	return nil
}

// SetGroupPermissionOnItem grants a group a permission on an item with source_group_id=group and origin="group_membership".
func (ctx *TestContext) SetGroupPermissionOnItem(permission, value, group, item string) error {
	ctx.setGrantedPermission(group, item, group, "group_membership", permission, value)

	return nil
}

// IHavePermissionOnItem gives the current user a permission on an item with source_group_id=user and origin="group_membership".
func (ctx *TestContext) IHavePermissionOnItem(permType, permValue, item string) error {
	return ctx.SetGroupPermissionOnItem(permType, permValue, ctx.user, item)
}

// GroupHasViewPermissionOnItem gives a group a "view" permission on an item with source_group_id=group and origin="group_membership".
func (ctx *TestContext) GroupHasViewPermissionOnItem(group, viewPermissionValue, item string) error {
	return ctx.SetGroupPermissionOnItem("can_view", viewPermissionValue, group, item)
}

// IHaveViewPermissionOnItem gives the current user a "view" permission on an item with source_group_id=user and origin="group_membership".
func (ctx *TestContext) IHaveViewPermissionOnItem(viewValue, item string) error {
	return ctx.SetGroupPermissionOnItem("can_view", viewValue, ctx.user, item)
}

// GroupHasGrantViewPermissionOnItem gives a group a "grant_view" permission on an item
// with source_group_id=group and origin="group_membership".
func (ctx *TestContext) GroupHasGrantViewPermissionOnItem(viewValue, group, item string) error {
	return ctx.SetGroupPermissionOnItem("can_grant_view", viewValue, group, item)
}

// GroupHasWatchPermissionOnItem gives a group a "watch" permission on an item with source_group_id=group and origin="group_membership".
func (ctx *TestContext) GroupHasWatchPermissionOnItem(group, watchValue, item string) error {
	return ctx.SetGroupPermissionOnItem("can_watch", watchValue, group, item)
}

// IHaveWatchPermissionOnItem gives the current user a "watch" permission on an item
// with source_group_id=user and origin="group_membership".
func (ctx *TestContext) IHaveWatchPermissionOnItem(watchValue, item string) error {
	return ctx.SetGroupPermissionOnItem("can_watch", watchValue, ctx.user, item)
}

// UserIsOwnerOfItem sets the is_owner permission.
func (ctx *TestContext) UserIsOwnerOfItem(isOwner, user, item string) error {
	return ctx.SetGroupPermissionOnItem("is_owner", isOwner, user, item)
}

// UserCanRequestHelpToOnItem sets the can_request_help_to permission.
func (ctx *TestContext) UserCanRequestHelpToOnItem(helperGroup, user, item string) error {
	return ctx.SetGroupPermissionOnItem("can_request_help_to", helperGroup, user, item)
}

// GroupHasValidatedResultOnItem adds a validated result of the group on the item to the database.
func (ctx *TestContext) GroupHasValidatedResultOnItem(group, item string) error {
	attemptID := rand.Int63()

	ctx.addAttempt(item, group)
	ctx.addValidatedResult(
		strconv.FormatInt(attemptID, 10),
		group,
		item,
		time.Now(),
	)

	return nil
}

// ThereAreTheFollowingValidatedResults creates validated results described in the given Godog table.
func (ctx *TestContext) ThereAreTheFollowingValidatedResults(results *godog.Table) error {
	for i := 1; i < len(results.Rows); i++ {
		result := ctx.getRowMap(i, results)

		ctx.addItem(map[string]string{
			"item": result["item"],
		})

		err := ctx.GroupHasValidatedResultOnItem(result["participant"], result["item"])
		if err != nil {
			return err
		}
	}

	return nil
}

// IHaveValidatedResultOnItem states that the current user has a validated result on the item.
func (ctx *TestContext) IHaveValidatedResultOnItem(item string) error {
	return ctx.GroupHasValidatedResultOnItem(ctx.user, item)
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
			err := ctx.GroupHasViewPermissionOnItem(thread["participant"], "content", thread["item"])
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
		helperGroupIDString := strconv.FormatInt(helperGroupID, 10)

		err := ctx.ThereIsAGroup(helperGroupIDString)
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
		thread["latest_update_at"] = time.Now().Format(time.DateTime)
	}

	ctx.currentThreadKey = ctx.getThreadKey(
		ctx.getIDOrIDByReference(thread["item_id"]),
		ctx.getIDOrIDByReference(thread["participant_id"]),
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
	rowIndex := ctx.getDBTableRowIndexForPrimaryKey("threads", ctx.currentThreadKey)
	if rowIndex == -1 {
		panic(fmt.Sprintf("IAmPartOfTheHelperGroupOfTheThread: thread %v not found", ctx.currentThreadKey))
	}
	columnIndex := getDBTableColumnIndex(ctx.dbTableData["threads"], "helper_group_id")
	if columnIndex == -1 {
		panic("IAmPartOfTheHelperGroupOfTheThread: helper_group_id column not found")
	}
	threadHelperGroupIDString := ctx.dbTableData["threads"].Rows[rowIndex].Cells[columnIndex].Value

	ctx.GroupIsAMemberOfTheGroup(ctx.user, threadHelperGroupIDString)

	return nil
}

// ICanRequestHelpToTheGroupWithIDOnTheItemWithID gives the user the permission to request help from a given group
// to a given item.
func (ctx *TestContext) ICanRequestHelpToTheGroupWithIDOnTheItemWithID(group, item string) error {
	return ctx.UserCanRequestHelpToOnItem(group, ctx.user, item)
}
