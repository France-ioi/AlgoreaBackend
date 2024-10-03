//go:build !prod

package testhelpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
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

func (ctx *TestContext) isInDatabase(tableName string, primaryKey map[string]string) bool {
	return ctx.getDBTableRowIndexForPrimaryKey(tableName, primaryKey) != -1
}

// addPersonalInfoViewApprovedFor adds a permission generated in the database.
func (ctx *TestContext) addPersonalInfoViewApprovedFor(childGroup, parentGroup string) {
	parentGroupID := ctx.getIDOfReference(parentGroup)
	childGroupID := ctx.getIDOfReference(childGroup)

	const groupGroupTable = "groups_groups"
	key := ctx.getGroupGroupKey(parentGroupID, childGroupID)
	if !ctx.isInDatabase(groupGroupTable, key) {
		ctx.addGroupGroup(parentGroup, childGroup)
	}

	ctx.setDBTableRowColumnValue(groupGroupTable, key, "personal_info_view_approved_at", time.Now().UTC().Format(time.DateTime))
}

// getGroupGroupKey gets a group group unique key for the groupgroup's map.
func (ctx *TestContext) getGroupGroupKey(parentGroupID, childGroupID int64) map[string]string {
	return map[string]string{
		"parent_group_id": strconv.FormatInt(parentGroupID, 10),
		"child_group_id":  strconv.FormatInt(childGroupID, 10),
	}
}

// addGroupGroup adds a group-group in the database.
func (ctx *TestContext) addGroupGroup(parentGroup, childGroup string) {
	parentGroupID := ctx.getIDOfReference(parentGroup)
	childGroupID := ctx.getIDOfReference(childGroup)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("groups_groups", &godog.Table{
		Rows: []*messages.PickleTableRow{
			{Cells: []*messages.PickleTableCell{{Value: "parent_group_id"}, {Value: "child_group_id"}}},
			{Cells: []*messages.PickleTableCell{{Value: strconv.FormatInt(parentGroupID, 10)}, {Value: strconv.FormatInt(childGroupID, 10)}}},
		},
	})
	if err != nil {
		panic(err)
	}
}

// addGroupManager adds a group manager in the database.
func (ctx *TestContext) addGroupManager(manager, group, canWatchMembers, canGrantGroupAccess, canManage string) {
	managerID := ctx.getIDOfReference(manager)
	groupID := ctx.getIDOfReference(group)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("group_managers", &godog.Table{
		Rows: []*messages.PickleTableRow{
			{Cells: []*messages.PickleTableCell{
				{Value: "manager_id"},
				{Value: "group_id"},
				{Value: "can_watch_members"},
				{Value: "can_grant_group_access"},
				{Value: "can_manage"},
			}},
			{Cells: []*messages.PickleTableCell{
				{Value: strconv.FormatInt(managerID, 10)},
				{Value: strconv.FormatInt(groupID, 10)},
				{Value: canWatchMembers},
				{Value: canGrantGroupAccess},
				{Value: canManage},
			}},
		},
	})
	if err != nil {
		panic(err)
	}
}

// addPermissionsGranted adds a permission granted in the database.
func (ctx *TestContext) addPermissionGranted(group, item, sourceGroup, origin, permission, permissionValue string) {
	groupIDString := strconv.FormatInt(ctx.getIDOfReference(group), 10)
	sourceGroupIDString := strconv.FormatInt(ctx.getIDOfReference(sourceGroup), 10)
	itemIDString := strconv.FormatInt(ctx.getIDOfReference(item), 10)

	const permissionsGrantedTable = "permissions_granted"
	primaryKey := map[string]string{
		"group_id": groupIDString, "item_id": itemIDString, "source_group_id": sourceGroupIDString, "origin": origin,
	}

	if !ctx.isInDatabase(permissionsGrantedTable, primaryKey) {
		ctx.needPopulateDatabase = true
		err := ctx.DBHasTable(permissionsGrantedTable, &godog.Table{
			Rows: []*messages.PickleTableRow{
				{Cells: []*messages.PickleTableCell{
					{Value: "group_id"}, {Value: "source_group_id"}, {Value: "item_id"}, {Value: "origin"},
				}},
				{Cells: []*messages.PickleTableCell{
					{Value: groupIDString}, {Value: sourceGroupIDString}, {Value: itemIDString}, {Value: origin},
				}},
			},
		})
		if err != nil {
			panic(err)
		}
	}

	if permission == "can_request_help_to" {
		canRequestHelpToGroupID := ctx.getIDOfReference(permissionValue)

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
	itemID := ctx.getIDOfReference(item)
	participantID := ctx.getIDOfReference(participant)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("attempts", &godog.Table{
		Rows: []*messages.PickleTableRow{
			{Cells: []*messages.PickleTableCell{{Value: "id"}, {Value: "participant_id"}}},
			{Cells: []*messages.PickleTableCell{{Value: strconv.FormatInt(itemID, 10)}, {Value: strconv.FormatInt(participantID, 10)}}},
		},
	})
	if err != nil {
		panic(err)
	}
}

// addValidatedResult adds a validated result to the database.
func (ctx *TestContext) addValidatedResult(attemptID, participant, item string, validatedAt time.Time) {
	participantID := ctx.getIDOfReference(participant)
	itemID := ctx.getIDOfReference(item)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("results", &godog.Table{
		Rows: []*messages.PickleTableRow{
			{Cells: []*messages.PickleTableCell{{Value: "attempt_id"}, {Value: "participant_id"}, {Value: "item_id"}, {Value: "validated_at"}}},
			{Cells: []*messages.PickleTableCell{
				{Value: attemptID},
				{Value: strconv.FormatInt(participantID, 10)},
				{Value: strconv.FormatInt(itemID, 10)},
				{Value: validatedAt.UTC().Format(time.DateTime)},
			}},
		},
	})
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
	parentItemID := ctx.getIDOfReference(parentItem)
	childItemID := ctx.getIDOfReference(childItem)

	ctx.needPopulateDatabase = true
	err := ctx.DBHasTable("items_items", &godog.Table{
		Rows: []*messages.PickleTableRow{
			{Cells: []*messages.PickleTableCell{{Value: "parent_item_id"}, {Value: "child_item_id"}, {Value: "child_order"}}},
			{Cells: []*messages.PickleTableCell{
				{Value: strconv.FormatInt(parentItemID, 10)},
				{Value: strconv.FormatInt(childItemID, 10)},
				{Value: strconv.FormatInt((parentItemID+childItemID)%1000, 10)},
			}},
		},
	})
	if err != nil {
		panic(err)
	}
}

func (ctx *TestContext) addItemItemPropagation(parent, child, propagation, propagationValue string) {
	primaryKey := ctx.getItemItemPrimaryKey(ctx.getIDOfReference(parent), ctx.getIDOfReference(child))
	ctx.setDBTableRowColumnValue("items_items", primaryKey, propagation, propagationValue)
}

// addItem adds an item in the database.
func (ctx *TestContext) addItem(fields map[string]string) {
	dbFields, primaryKey, oldRowIndex := ctx.constructDBFieldsForAddItem(fields)

	ctx.needPopulateDatabase = true

	if oldRowIndex != -1 {
		ctx.setDBTableRowColumnValues("items", primaryKey, dbFields)
		return
	}

	columnNameCells := make([]*messages.PickleTableCell, 0, len(dbFields))
	columnValues := make([]*messages.PickleTableCell, 0, len(dbFields))
	for key, value := range dbFields {
		columnNameCells = append(columnNameCells, &messages.PickleTableCell{Value: key})
		columnValues = append(columnValues, &messages.PickleTableCell{Value: value})
	}
	err := ctx.DBHasTable("items", &godog.Table{
		Rows: []*messages.PickleTableRow{
			{Cells: columnNameCells}, {Cells: columnValues},
		},
	})
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
			key = "id"
		}

		switch {
		case strings.HasSuffix(key, "id"):
			dbFields[key] = strconv.FormatInt(ctx.getIDOfReference(value), 10)
		case value[0] == referencePrefix:
			dbFields[key] = value[1:]
		default:
			dbFields[key] = value
		}
	}

	primaryKey = map[string]string{"id": dbFields["id"]}
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
	itemID := ctx.getIDOfReference(item)
	participantID := ctx.getIDOfReference(participant)
	helperGroupID := ctx.getIDOfReference(helperGroup)

	_, err := time.Parse(time.DateTime, latestUpdateAt)
	if err != nil {
		panic(err)
	}

	err = ctx.DBHasTable("threads", &godog.Table{
		Rows: []*messages.PickleTableRow{
			{Cells: []*messages.PickleTableCell{
				{Value: "item_id"},
				{Value: "participant_id"},
				{Value: "helper_group_id"},
				{Value: "status"},
				{Value: "message_count"},
				{Value: "latest_update_at"},
			}},
			{Cells: []*messages.PickleTableCell{
				{Value: strconv.FormatInt(itemID, 10)},
				{Value: strconv.FormatInt(participantID, 10)},
				{Value: strconv.FormatInt(helperGroupID, 10)},
				{Value: status},
				{Value: messageCount},
				{Value: latestUpdateAt},
			}},
		},
	})
	if err != nil {
		panic(err)
	}
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

func (ctx *TestContext) userHasValidatedResultOnItem(user, item string) error {
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

		err := ctx.userHasValidatedResultOnItem(result["participant"], result["item"])
		if err != nil {
			return err
		}
	}

	return nil
}

// IHaveValidatedResultOnItem states that the current user has a validated result on the item.
func (ctx *TestContext) IHaveValidatedResultOnItem(item string) error {
	return ctx.userHasValidatedResultOnItem(ctx.user, item)
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
	rowIndex := ctx.getDBTableRowIndexForPrimaryKey("threads", ctx.currentThreadKey)
	if rowIndex == -1 {
		panic(fmt.Sprintf("IAmPartOfTheHelperGroupOfTheThread: thread %v not found", ctx.currentThreadKey))
	}
	columnIndex := getDBTableColumnIndex(ctx.dbTableData["threads"], "helper_group_id")
	if columnIndex == -1 {
		panic("IAmPartOfTheHelperGroupOfTheThread: helper_group_id column not found")
	}
	threadHelperGroupIDString := ctx.dbTableData["threads"].Rows[rowIndex].Cells[columnIndex].Value

	ctx.IsAMemberOfTheGroup(ctx.user, threadHelperGroupIDString)

	return nil
}

// ICanRequestHelpToTheGroupWithIDOnTheItemWithID gives the user the permission to request help from a given group
// to a given item.
func (ctx *TestContext) ICanRequestHelpToTheGroupWithIDOnTheItemWithID(group, item string) error {
	return ctx.UserCanRequestHelpToOnItemWithID(group, ctx.user, item)
}
