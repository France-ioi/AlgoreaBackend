//go:build !prod

package testhelpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/messages-go/v10"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/rand"
	"github.com/France-ioi/AlgoreaBackend/app/utils"
)

const (
	ReferencePrefix = '@'
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
	if reference[0] == ReferencePrefix {
		return reference[1:]
	}

	return reference
}

// getRowMap convert a PickleTable's row into a map where the keys are the column headers.
func (ctx *TestContext) getRowMap(rowIndex int, table *messages.PickleStepArgument_PickleTable) map[string]string {
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

	db, err := database.Open(ctx.db())
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

func (ctx *TestContext) addInDatabase(tableName, key string, row map[string]interface{}) {
	ctx.needPopulateDatabase = true

	if ctx.dbTables[tableName] == nil {
		ctx.dbTables[tableName] = make(map[string]map[string]interface{})
	}

	ctx.dbTables[tableName][key] = row
}

// addUsersIntoAllUsersGroup adds all users in the AllUsers group if it is defined.
func (ctx *TestContext) addUsersIntoAllUsersGroup() error {
	if ctx.allUsersGroup == "" {
		return nil
	}

	for userID := range ctx.dbTables["users"] {
		err := ctx.UserIsAMemberOfTheGroup(userID, ctx.allUsersGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

// addUser adds a user in database.
func (ctx *TestContext) addUser(fields map[string]string) {
	dbFields := make(map[string]interface{})
	for key, value := range fields {
		if key == "user" {
			key = "login"
		}

		switch {
		case strings.HasSuffix(key, "_id"):
			dbFields[key] = ctx.getReference(value)
		case value[0] == ReferencePrefix:
			dbFields[key] = value[1:]
		default:
			dbFields[key] = value
		}
	}

	userKey := strconv.FormatInt(dbFields["group_id"].(int64), 10)

	if oldFields, ok := ctx.dbTables["users"][userKey]; ok {
		dbFields = mergeFields(oldFields, dbFields)
	}

	ctx.addInDatabase("users", userKey, dbFields)
}

// addGroup adds a group in database.
func (ctx *TestContext) addGroup(group, name, groupType string) {
	groupID := ctx.getReference(group)

	ctx.addInDatabase("groups", strconv.FormatInt(groupID, 10), map[string]interface{}{
		"id":   groupID,
		"name": referenceToName(name),
		"type": groupType,
	})
}

// addPersonalInfoViewApprovedFor adds a permission generated in the database.
func (ctx *TestContext) addPersonalInfoViewApprovedFor(childGroup, parentGroup string) {
	parentGroupID := ctx.getReference(parentGroup)
	childGroupID := ctx.getReference(childGroup)

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
	parentGroupID := ctx.getReference(parentGroup)
	childGroupID := ctx.getReference(childGroup)

	ctx.addInDatabase("groups_groups", ctx.getGroupGroupKey(parentGroupID, childGroupID), map[string]interface{}{
		"parent_group_id": parentGroupID,
		"child_group_id":  childGroupID,
	})
}

// addGroupManager adds a group manager in the database.
func (ctx *TestContext) addGroupManager(manager, group, canWatchMembers, canGrantGroupAccess, canManage string) {
	managerID := ctx.getReference(manager)
	groupID := ctx.getReference(group)

	ctx.addInDatabase(
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
	groupID := ctx.getReference(group)
	sourceGroupID := ctx.getReference(sourceGroup)
	itemID := ctx.getReference(item)

	permissionsGrantedTable := "permissions_granted"
	key := strconv.FormatInt(groupID, 10) + "," +
		strconv.FormatInt(itemID, 10) + "," +
		strconv.FormatInt(sourceGroupID, 10) + "," +
		origin

	if !ctx.isInDatabase(permissionsGrantedTable, key) {
		ctx.addInDatabase(permissionsGrantedTable, key, map[string]interface{}{
			"group_id":        groupID,
			"source_group_id": sourceGroupID,
			"item_id":         itemID,
			"origin":          origin,
		})
	}

	if permission == "can_request_help_to" {
		canRequestHelpToGroupID := strconv.FormatInt(ctx.getReference(permissionValue), 10)

		if !ctx.isInDatabase("groups", canRequestHelpToGroupID) {
			ctx.addGroup(permissionValue, "Group "+referenceToName(permissionValue), "Class")
		}

		permissionValue = canRequestHelpToGroupID
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

// addAttempt adds an attempt in database.
func (ctx *TestContext) addAttempt(item, participant string) {
	itemID := ctx.getReference(item)
	participantID := ctx.getReference(participant)

	ctx.addInDatabase(
		`attempts`,
		strconv.FormatInt(itemID, 10)+","+strconv.FormatInt(participantID, 10),
		map[string]interface{}{
			"id":             itemID,
			"participant_id": participantID,
		},
	)
}

// addResult adds a result in database.
func (ctx *TestContext) addResult(attemptID, participant, item string, validatedAt time.Time) {
	participantID := ctx.getReference(participant)
	itemID := ctx.getReference(item)

	ctx.addInDatabase(
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
	parentItemID := ctx.getReference(parentItem)
	childItemID := ctx.getReference(childItem)

	ctx.addInDatabase(
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
	key := ctx.getItemItemKey(ctx.getReference(parent), ctx.getReference(child))

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
			dbFields[key] = ctx.getReference(value)
		case value[0] == ReferencePrefix:
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
	if _, ok := dbFields["text_id"]; !ok && fields["item"][0] == ReferencePrefix {
		dbFields["text_id"] = fields["item"][1:]
	}

	ctx.addInDatabase("items", itemKey, dbFields)
}

// getThreadKey gets a thread unique key for the thread's map.
func (ctx *TestContext) getThreadKey(itemID, participantID int64) string {
	return strconv.FormatInt(itemID, 10) + "," + strconv.FormatInt(participantID, 10)
}

// addThread adds a thread in database.
func (ctx *TestContext) addThread(item, participant, helperGroup, status, messageCount, latestUpdateAt string) {
	itemID := ctx.getReference(item)
	participantID := ctx.getReference(participant)
	helperGroupID := ctx.getReference(helperGroup)

	latestUpdateAtDate, err := time.Parse(utils.DateTimeFormat, latestUpdateAt)
	if err != nil {
		panic(err)
	}

	ctx.addInDatabase("threads", ctx.getThreadKey(itemID, participantID), map[string]interface{}{
		"item_id":          itemID,
		"participant_id":   participantID,
		"helper_group_id":  helperGroupID,
		"status":           status,
		"message_count":    messageCount,
		"latest_update_at": latestUpdateAtDate,
	})
}

// addSession adds a session in database.
func (ctx *TestContext) addSession(session, user, refreshToken string) {
	sessionID := ctx.getReference(session)
	userID := ctx.getReference(user)

	ctx.addInDatabase("sessions", strconv.FormatInt(sessionID, 10), map[string]interface{}{
		"session_id":    sessionID,
		"user_id":       userID,
		"refresh_token": refreshToken,
	})
}

// addAccessToken adds a access token in database.
func (ctx *TestContext) addAccessToken(session, token, issuedAt, expiresAt string) {
	sessionID := ctx.getReference(session)

	issuedAtDate, err := time.Parse(utils.DateTimeFormat, issuedAt)
	if err != nil {
		panic(err)
	}

	expiresAtDate, err := time.Parse(utils.DateTimeFormat, expiresAt)
	if err != nil {
		panic(err)
	}

	ctx.addInDatabase("access_tokens", token, map[string]interface{}{
		"session_id": sessionID,
		"token":      token,
		"issued_at":  issuedAtDate,
		"expires_at": expiresAtDate,
	})
}

// IAm Sets the current user.
func (ctx *TestContext) IAm(name string) error {
	err := ctx.ThereIsAUser(name)
	if err != nil {
		return err
	}

	err = ctx.IAmUserWithID(ctx.getReference(name))
	if err != nil {
		return err
	}

	ctx.user = name

	return nil
}

// ThereIsAUser create a user.
func (ctx *TestContext) ThereIsAUser(name string) error {
	return ctx.ThereIsAUserWith(getParameterString(map[string]string{
		"group_id": name,
		"user":     name,
	}))
}

// ThereIsAUserWith creates a new user.
func (ctx *TestContext) ThereIsAUserWith(parameters string) error {
	user := ctx.getParameterMap(parameters)

	if _, ok := user["group_id"]; !ok {
		user["group_id"] = user["user"]
	}

	ctx.addUser(user)

	return ctx.ThereIsAGroupWith(getParameterString(map[string]string{
		"id":   user["group_id"],
		"name": user["user"],
		"type": "User",
	}))
}

// ThereAreTheFollowingGroups defines groups.
func (ctx *TestContext) ThereAreTheFollowingGroups(groups *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(groups.Rows); i++ {
		group := ctx.getRowMap(i, groups)

		err := ctx.ThereIsAGroup(group["group"])
		if err != nil {
			return err
		}

		if _, ok := group["parent"]; ok {
			err = ctx.GroupIsAChildOfTheGroup(group["group"], group["parent"])
			if err != nil {
				return err
			}
		}

		if _, ok := group["members"]; ok {
			members := strings.Split(group["members"], ",")

			for _, member := range members {
				err = ctx.ThereIsAUser(member)
				if err != nil {
					return err
				}

				err = ctx.GroupIsAChildOfTheGroup(member, group["group"])
				if err != nil {
					return err
				}

				err = ctx.ThereIsAGroupWith(getParameterString(map[string]string{
					"id":   member,
					"name": member,
					"type": "User",
				}))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ThereIsAGroupWith creates a new group.
func (ctx *TestContext) ThereIsAGroupWith(parameters string) error {
	group := ctx.getParameterMap(parameters)

	if _, ok := group["name"]; !ok {
		group["name"] = "Group " + referenceToName(group["id"])
	}
	if _, ok := group["type"]; !ok {
		group["type"] = "Class"
	}

	ctx.addGroup(group["id"], group["name"], group["type"])

	return nil
}

// ThereAreTheFollowingUsers defines users.
func (ctx *TestContext) ThereAreTheFollowingUsers(users *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(users.Rows); i++ {
		user := ctx.getRowMap(i, users)

		err := ctx.ThereIsAUserWith(getParameterString(user))
		if err != nil {
			return err
		}

		err = ctx.ThereIsAGroup(user["user"])
		if err != nil {
			return err
		}
	}

	return nil
}

// ThereIsAGroup creates a new group.
func (ctx *TestContext) ThereIsAGroup(group string) error {
	return ctx.ThereIsAGroupWith(getParameterString(map[string]string{
		"id":   group,
		"name": group,
	}))
}

// UserIsAManagerOfTheGroupWith sets the current user as the manager of a group.
func (ctx *TestContext) UserIsAManagerOfTheGroupWith(parameters string) error {
	err := ctx.ThereIsAGroupWith(parameters)
	if err != nil {
		return err
	}

	// We create a parent group of which the user is the manager.
	group := ctx.getParameterMap(parameters)

	canWatchMembers := "0"
	canGrantGroupAccess := "0"
	canManage := "none"
	watchedGroupName := group["user_id"] + " manages " + referenceToName(group["name"])

	if group["can_watch_members"] == strTrue {
		canWatchMembers = "1"
		watchedGroupName += " with can_watch_members"
	}
	if group["can_grant_group_access"] == strTrue {
		canGrantGroupAccess = "1"
		watchedGroupName += " with can_grant_group_access"
	}
	if _, ok := group["can_manage"]; ok {
		canManage = group["can_manage"]
		watchedGroupName += " with can_manage_memberships_and_groups"
	}

	err = ctx.ThereIsAGroupWith(getParameterString(map[string]string{
		"id":   watchedGroupName,
		"name": watchedGroupName,
	}))
	if err != nil {
		return err
	}

	ctx.IsAMemberOfTheGroup(group["id"], watchedGroupName)

	ctx.addGroupManager(group["user_id"], watchedGroupName, canWatchMembers, canGrantGroupAccess, canManage)

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

// UserIsAManagerOfTheGroupAndCanWatchItsMembers sets the user as a manager of a group with can_watch permission.
func (ctx *TestContext) UserIsAManagerOfTheGroupAndCanWatchItsMembers(user, group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                group,
		"user_id":           user,
		"name":              group,
		"can_watch_members": strTrue,
	}))
}

func (ctx *TestContext) UserIsAManagerOfTheGroupAndCanGrantGroupAccess(user, group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":                     group,
		"user_id":                user,
		"name":                   group,
		"can_grant_group_access": strTrue,
	}))
}

// UserIsAManagerOfTheGroupAndCanManageMembershipsAndGroup adds a user as a manager of a group
// with the can_manage=memberships_and_groups permission.
func (ctx *TestContext) UserIsAManagerOfTheGroupAndCanManageMembershipsAndGroup(user, group string) error {
	return ctx.UserIsAManagerOfTheGroupWith(getParameterString(map[string]string{
		"id":         group,
		"user_id":    user,
		"name":       group,
		"can_manage": "memberships_and_group",
	}))
}

// theGroupIsADescendantOfTheGroup sets a group as a descendant of another.
func (ctx *TestContext) theGroupIsADescendantOfTheGroup(descendant, parent string) error {
	// we add another group in between to increase the robustness of the tests.
	middle := parent + " -> X -> " + referenceToName(descendant)

	groups := []string{descendant, middle, parent}
	for _, group := range groups {
		err := ctx.ThereIsAGroupWith(getParameterString(map[string]string{
			"id": group,
		}))
		if err != nil {
			return err
		}
	}

	ctx.IsAMemberOfTheGroup(middle, parent)
	ctx.IsAMemberOfTheGroup(descendant, middle)

	return nil
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
func (ctx *TestContext) ThereAreTheFollowingItems(items *messages.PickleStepArgument_PickleTable) error {
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
func (ctx *TestContext) ThereAreTheFollowingTasks(tasks *messages.PickleStepArgument_PickleTable) error {
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
func (ctx *TestContext) ThereAreTheFollowingItemPermissions(itemPermissions *messages.PickleStepArgument_PickleTable) error {
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
func (ctx *TestContext) ThereAreTheFollowingItemRelations(itemPermissions *messages.PickleStepArgument_PickleTable) error {
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

// IAmAMemberOfTheGroupWithID creates a group and add the user in it.
func (ctx *TestContext) IAmAMemberOfTheGroupWithID(group string) error {
	err := ctx.ThereIsAGroupWith("id=" + group)
	if err != nil {
		return err
	}

	ctx.IsAMemberOfTheGroup(
		ctx.user,
		group,
	)

	return nil
}

// GroupIsAChildOfTheGroup puts a group as a child of another group.
func (ctx *TestContext) GroupIsAChildOfTheGroup(childGroup, parentGroup string) error {
	err := ctx.ThereIsAGroupWith(getParameterString(map[string]string{
		"id":   childGroup,
		"name": childGroup,
	}))
	if err != nil {
		return err
	}

	err = ctx.ThereIsAGroupWith(getParameterString(map[string]string{
		"id":   parentGroup,
		"name": parentGroup,
	}))
	if err != nil {
		return err
	}

	ctx.IsAMemberOfTheGroup(childGroup, parentGroup)

	return nil
}

// UserIsAMemberOfTheGroup puts a user in a group.
func (ctx *TestContext) UserIsAMemberOfTheGroup(user, group string) error {
	err := ctx.ThereIsAUserWith(getParameterString(map[string]string{
		"group_id": user,
		"user":     user,
	}))
	if err != nil {
		return err
	}

	return ctx.GroupIsAChildOfTheGroup(user, group)
}

// UserIsAMemberOfTheGroupWhoHasApprovedAccessToHisPersonalInfo puts a user in a group with approved access to his personnel info.
func (ctx *TestContext) UserIsAMemberOfTheGroupWhoHasApprovedAccessToHisPersonalInfo(user, group string) error {
	err := ctx.UserIsAMemberOfTheGroup(user, group)
	if err != nil {
		return err
	}

	ctx.addPersonalInfoViewApprovedFor(user, group)

	return nil
}

// AllUsersGroupIsDefinedAsTheGroup creates and sets the allUsersGroup.
func (ctx *TestContext) AllUsersGroupIsDefinedAsTheGroup(group string) error {
	err := ctx.ThereIsAGroupWith(getParameterString(map[string]string{
		"id":   group,
		"name": "AllUsers",
		"type": "Base",
	}))
	if err != nil {
		return err
	}

	err = ctx.TheApplicationConfigIs(&messages.PickleStepArgument_PickleDocString{
		Content: `
domains:
  -
    domains: [127.0.0.1]
    allUsersGroup: ` + group + `
`,
	})
	if err != nil {
		return err
	}

	ctx.allUsersGroup = group

	return nil
}

// TheFieldOfTheGroupShouldBe checks that the field of a group in the database is equal to a value.
func (ctx *TestContext) TheFieldOfTheGroupShouldBe(field, group, value string) error {
	groupID := ctx.getReference(group)

	var resultCount int
	err := db.QueryRow("SELECT COUNT(*) as count FROM `groups` WHERE id = ? AND "+field+" = ?", groupID, value).
		Scan(&resultCount)
	if err != nil {
		return err
	}

	if resultCount != 1 {
		return fmt.Errorf("expected the group %s have %s=%s", group, field, value)
	}

	return nil
}

// UserShouldNotBeAMemberOfTheGroup checks that the user is not a member of the group.
func (ctx *TestContext) UserShouldNotBeAMemberOfTheGroup(user, group string) error {
	userID := ctx.getReference(user)
	groupID := ctx.getReference(group)

	var resultCount int
	err := db.QueryRow("SELECT COUNT(*) as count FROM `groups_groups` WHERE parent_group_id = ? AND child_group_id = ?", groupID, userID).
		Scan(&resultCount)
	if err != nil {
		return err
	}

	if resultCount != 0 {
		return fmt.Errorf("expected the user %s not to be a member of the group %s", user, group)
	}

	return nil
}

// UserShouldBeAMemberOfTheGroup checks that the user is a member of the group.
func (ctx *TestContext) UserShouldBeAMemberOfTheGroup(user, group string) error {
	userID := ctx.getReference(user)
	groupID := ctx.getReference(group)

	var resultCount int
	err := db.QueryRow("SELECT COUNT(*) as count FROM `groups_groups` WHERE parent_group_id = ? AND child_group_id = ?", groupID, userID).
		Scan(&resultCount)
	if err != nil {
		return err
	}

	if resultCount != 1 {
		return fmt.Errorf("expected the user %s to be a member of the group %s", user, group)
	}

	return nil
}

// IAmAMemberOfTheGroup puts a user in a group.
func (ctx *TestContext) IAmAMemberOfTheGroup(name string) error {
	return ctx.IAmAMemberOfTheGroupWithID(name)
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

func (ctx *TestContext) UserHaveValidatedItemWithID(user, item string) error {
	attemptID := rand.Int63()

	ctx.addAttempt(item, user)
	ctx.addResult(
		strconv.FormatInt(attemptID, 10),
		user,
		item,
		time.Now(),
	)

	return nil
}

func (ctx *TestContext) ThereAreTheFollowingResults(results *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(results.Rows); i++ {
		result := ctx.getRowMap(i, results)

		ctx.addItem(map[string]string{
			"item": result["item"],
		})

		err := ctx.UserHaveValidatedItemWithID(result["participant"], result["item"])
		if err != nil {
			return err
		}
	}

	return nil
}

// IHaveValidatedItemWithID states that user has validated an item.
func (ctx *TestContext) IHaveValidatedItemWithID(item string) error {
	return ctx.UserHaveValidatedItemWithID(ctx.user, item)
}

// ThereAreTheFollowingThreads create threads.
func (ctx *TestContext) ThereAreTheFollowingThreads(threads *messages.PickleStepArgument_PickleTable) error {
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

		err := ctx.ThereIsAGroupWith(getParameterString(map[string]string{
			"id":   strconv.FormatInt(helperGroupID, 10),
			"name": "helper_group_for_" + thread["item_id"] + "-" + thread["participant_id"],
		}))
		if err != nil {
			return err
		}

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
		ctx.getReference(thread["item_id"]),
		ctx.getReference(thread["participant_id"]),
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

// ThereAreTheFollowingSessions create sessions.
func (ctx *TestContext) ThereAreTheFollowingSessions(sessions *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(sessions.Rows); i++ {
		session := ctx.getRowMap(i, sessions)

		ctx.addSession(
			session["session"],
			session["user"],
			session["refresh_token"],
		)
	}

	return nil
}

// ThereAreTheFollowingAccessTokens create access tokens.
func (ctx *TestContext) ThereAreTheFollowingAccessTokens(accessTokens *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(accessTokens.Rows); i++ {
		accessToken := ctx.getRowMap(i, accessTokens)

		ctx.addAccessToken(
			accessToken["session"],
			accessToken["token"],
			accessToken["issued_at"],
			accessToken["expires_at"],
		)
	}

	return nil
}
