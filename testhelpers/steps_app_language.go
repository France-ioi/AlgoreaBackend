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
)

const (
	ReferenceHeader = "@reference"
	ReferencePrefix = '@'
)

// ctx.getParametersMap parses parameters in format key1=val1,key2=val2,... into a map.
func (ctx *TestContext) getParametersMap(parameters string) map[string]string {
	parametersMap := make(map[string]string)
	arrayParameters := strings.Split(parameters, ",")
	for _, paramKeyValue := range arrayParameters {
		keyVal := strings.Split(paramKeyValue, "=")
		if keyVal[1][0] == ReferencePrefix && keyVal[0] != ReferenceHeader {
			keyVal[1] = ctx.replaceReferencesByIDs(keyVal[1])
		}

		parametersMap[keyVal[0]] = keyVal[1]
	}

	return parametersMap
}

// getParametersString converts parameters into a string with format key1=val1,key2=val2,...
func getParametersString(parameters map[string]string) string {
	var str string
	for key, value := range parameters {
		if str != "" {
			str += ","
		}
		str += key + "=" + value
	}

	return str
}

// getRowMap convert a PickleTable's row into a map where the keys are the column headers.
func (ctx *TestContext) getRowMap(rowIndex int, table *messages.PickleStepArgument_PickleTable) map[string]string {
	rowHeader := table.Rows[0]
	sourceRow := table.Rows[rowIndex]

	rowMap := map[string]string{}
	for i := 0; i < len(rowHeader.Cells); i++ {
		value := sourceRow.Cells[i].Value

		// when the value in the cell is a reference to an identifier.
		if value[0] == ReferencePrefix && rowHeader.Cells[i].Value != ReferenceHeader {
			value = strconv.FormatInt(ctx.getOrCreateReferenceFor(value[1:]), 10)
		}

		rowMap[rowHeader.Cells[i].Value] = value
	}

	return rowMap
}

// populateDatabase populate the database with all the initialized data.
func (ctx *TestContext) populateDatabase() error {
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}

	// add all the defined table rows in the database.
	err = database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		for tableName, tableRows := range ctx.dbTables {
			for _, tableRow := range tableRows {
				SQLRow := make(map[string]interface{})
				for field, value := range tableRow {
					if field != ReferenceHeader {
						SQLRow[field] = value
					}
				}

				err = database.NewDataStoreWithTable(store.DB, tableName).InsertOrUpdateMap(SQLRow, nil)
				if err != nil {
					return fmt.Errorf("populateDatabase %s %+v: %v", tableName, SQLRow, err)
				}
			}
		}

		return nil
	})
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

func mergeFields(row1, row2 map[string]interface{}) map[string]interface{} {
	merged := row1
	for key, value := range row2 {
		merged[key] = value
	}

	return merged
}

func (ctx *TestContext) addInDatabase(tableName, key string, row map[string]interface{}) {
	if ctx.dbTables[tableName] == nil {
		ctx.dbTables[tableName] = make(map[string]map[string]interface{})
	}

	if oldRow, ok := ctx.dbTables[tableName][key]; ok {
		row = mergeFields(oldRow, row)
	}

	ctx.dbTables[tableName][key] = row
}

func (ctx *TestContext) getUserKey(fields map[string]string) string {
	if _, ok := fields["group_id"]; !ok {
		panic(fmt.Errorf("getUserKey: %v must have a group_id", fields))
	}

	return fields["group_id"]
}

// addUser adds a user in database.
func (ctx *TestContext) addUser(fields map[string]string) {
	dbFields := make(map[string]interface{})
	for key, value := range fields {
		dbFields[key] = value
	}

	ctx.addInDatabase("users", ctx.getUserKey(fields), dbFields)
}

// addGroup adds a group in database.
func (ctx *TestContext) addGroup(groupID, name, groupType string) {
	ctx.addInDatabase("groups", groupID, map[string]interface{}{
		"id":   groupID,
		"name": name,
		"type": groupType,
	})
}

// addPermissionGenerated adds a permission generated in database.
func (ctx *TestContext) addPersonalInfoViewApprovedFor(childGroupID, parentGroupID string) {
	groupGroupTable := "groups_groups"
	key := parentGroupID + "," + childGroupID
	if !ctx.isInDatabase(groupGroupTable, key) {
		ctx.addGroupGroup(parentGroupID, childGroupID)
	}

	ctx.dbTables[groupGroupTable][key]["personal_info_view_approved_at"] = time.Now()
}

// addGroupGroup adds a group-group in database.
func (ctx *TestContext) addGroupGroup(parentGroupID, childGroupID string) {
	ctx.addInDatabase("groups_groups", parentGroupID+","+childGroupID, map[string]interface{}{
		"parent_group_id": parentGroupID,
		"child_group_id":  childGroupID,
	})
}

// addGroupManager adds a group manager in database.
func (ctx *TestContext) addGroupManager(managerID, groupID, canWatchMembers string) {
	ctx.addInDatabase("group_managers", managerID+","+groupID, map[string]interface{}{
		"manager_id":        managerID,
		"group_id":          groupID,
		"can_watch_members": canWatchMembers,
	})
}

// addPermissionGenerated adds a permission generated in database.
func (ctx *TestContext) addPermissionGenerated(groupID, itemID, watchType, watchValue string) {
	permissionsGeneratedTable := "permissions_generated"
	key := groupID + "," + itemID
	if !ctx.isInDatabase(permissionsGeneratedTable, key) {
		ctx.addInDatabase(permissionsGeneratedTable, key, map[string]interface{}{
			"group_id": groupID,
			"item_id":  itemID,
		})
	}

	ctx.dbTables[permissionsGeneratedTable][key]["can_"+watchType+"_generated"] = watchValue
}

// addPermissionsGranted adds a permission granted in database.
func (ctx *TestContext) addPermissionGranted(groupID, sourceGroupID, itemID, canRequestHelpTo string) {
	ctx.addInDatabase("permissions_granted", groupID+","+itemID, map[string]interface{}{
		"group_id":            groupID,
		"source_group_id":     sourceGroupID,
		"item_id":             itemID,
		"can_request_help_to": canRequestHelpTo,
	})
}

// addAttempt adds an attempt in database.
func (ctx *TestContext) addAttempt(id, participantID string) {
	ctx.addInDatabase("attempts", id+","+participantID, map[string]interface{}{
		"id":             id,
		"participant_id": participantID,
	})
}

// addResult adds a result in database.
func (ctx *TestContext) addResult(attemptID, participantID, itemID string, validatedAt time.Time) {
	ctx.addInDatabase("results", attemptID+","+participantID+","+itemID, map[string]interface{}{
		"attempt_id":     attemptID,
		"participant_id": participantID,
		"item_id":        itemID,
		"validated_at":   validatedAt,
	})
}

// addItem adds an item in database.
func (ctx *TestContext) addItem(id, defaultLanguageTag, itemType string) {
	ctx.addInDatabase("items", id, map[string]interface{}{
		"id":                   id,
		"default_language_tag": defaultLanguageTag,
		"type":                 itemType,
	})
}

// getThreadKey gets a thread unique key for the threads map.
func (ctx *TestContext) getThreadKey(itemID, participantID string) string {
	return itemID + "," + participantID
}

// addThread adds a thread in database.
func (ctx *TestContext) addThread(itemID, participantID, helperGroupID, status, messageCount string) {
	ctx.addInDatabase("threads", ctx.getThreadKey(itemID, participantID), map[string]interface{}{
		"item_id":         itemID,
		"participant_id":  participantID,
		"helper_group_id": helperGroupID,
		"status":          status,
		"message_count":   messageCount,
	})
}

// IAm Sets the current user.
func (ctx *TestContext) IAm(name string) error {
	err := ctx.ThereIsAUser(name)
	if err != nil {
		return err
	}

	return ctx.IAmUserWithID(ctx.getOrCreateReferenceFor(name))
}

// ThereIsAUser create a user.
func (ctx *TestContext) ThereIsAUser(name string) error {
	userID := ctx.getOrCreateReferenceFor(name)

	return ctx.ThereIsAUserWith(getParametersString(map[string]string{
		"group_id": strconv.FormatInt(userID, 10),
		"login":    name,
	}))
}

// ThereIsAUserReferencedBy create a user with a reference.
func (ctx *TestContext) ThereIsAUserReferencedBy(name, reference string) error {
	userID := rand.Int63()

	return ctx.ThereIsAUserWith(getParametersString(map[string]string{
		ReferenceHeader: reference,
		"group_id":      strconv.FormatInt(userID, 10),
		"login":         name,
	}))
}

// ThereIsAUserWith creates a new user.
func (ctx *TestContext) ThereIsAUserWith(parameters string) error {
	user := ctx.getParametersMap(parameters)

	if _, ok := user["group_id"]; !ok {
		user["group_id"] = strconv.FormatInt(rand.Int63(), 10)
	}

	ctx.addUser(user)

	if reference, ok := user[ReferenceHeader]; ok {
		ctx.setReference(reference, user["group_id"])
	}

	return ctx.ThereIsAGroupWith(getParametersString(map[string]string{
		"id":   user["group_id"],
		"name": user["login"],
		"type": "User",
	}))
}

// ThereIsAGroupWith creates a new group.
func (ctx *TestContext) ThereIsAGroupWith(parameters string) error {
	group := ctx.getParametersMap(parameters)

	if _, ok := group["name"]; !ok {
		group["name"] = "Group " + group["id"]
	}
	if _, ok := group["type"]; !ok {
		group["type"] = "Class"
	}

	ctx.addGroup(group["id"], group["name"], group["type"])

	if reference, ok := group[ReferenceHeader]; ok {
		ctx.setReference(reference, group["id"])
	}

	return nil
}

// ThereAreTheFollowingUsers defines users.
func (ctx *TestContext) ThereAreTheFollowingUsers(users *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(users.Rows); i++ {
		user := ctx.getRowMap(i, users)

		err := ctx.ThereIsAUserWith(getParametersString(user))
		if err != nil {
			return err
		}

		err = ctx.ThereIsAGroup(user["login"])
		if err != nil {
			return err
		}
	}

	return nil
}

// ThereIsAGroup creates a new group.
func (ctx *TestContext) ThereIsAGroup(groupName string) error {
	groupID := strconv.FormatInt(ctx.getOrCreateReferenceFor(groupName), 10)

	return ctx.ThereIsAGroupWith(getParametersString(map[string]string{
		"id":   groupID,
		"name": groupName,
	}))
}

// ThereIsAGroupReferencedBy creates a new group and assigns it a reference.
func (ctx *TestContext) ThereIsAGroupReferencedBy(groupName, reference string) error {
	groupID := rand.Int63()

	return ctx.ThereIsAGroupWith(getParametersString(map[string]string{
		ReferenceHeader: reference,
		"id":            strconv.FormatInt(groupID, 10),
		"name":          groupName,
	}))
}

// IAmTheManagerOfTheGroupWith sets the current user as the manager of a group.
func (ctx *TestContext) IAmTheManagerOfTheGroupWith(parameters string) error {
	err := ctx.ThereIsAGroupWith(parameters)
	if err != nil {
		return err
	}

	// We create a parent group of which the user is the manager.
	group := ctx.getParametersMap(parameters)

	canWatchMembersValue := "0"
	watchedGroupName := "User manages " + group["name"]

	if group["can_watch_members"] == "true" {
		canWatchMembersValue = "1"
		watchedGroupName += " with can_watch_members"
	}

	watchedGroupID := rand.Int63()
	err = ctx.ThereIsAGroupWith(getParametersString(map[string]string{
		"id":   strconv.FormatInt(watchedGroupID, 10),
		"name": watchedGroupName,
	}))
	if err != nil {
		return err
	}

	groupID, _ := strconv.ParseInt(group["id"], 10, 64)
	ctx.IsAMemberOfTheGroup(groupID, watchedGroupID)

	ctx.addGroupManager(strconv.FormatInt(ctx.userID, 10), strconv.FormatInt(watchedGroupID, 10), canWatchMembersValue)

	return nil
}

// IAmTheManagerOfTheGroupWithID sets the user as a manager of a group with an id.
func (ctx *TestContext) IAmTheManagerOfTheGroupWithID(groupID int64) error {
	return ctx.IAmTheManagerOfTheGroupWith(getParametersString(map[string]string{
		"id":                strconv.FormatInt(groupID, 10),
		"can_watch_members": "false",
	}))
}

// IAmTheManagerOfTheGroup sets the user as a manager of a group with an id.
func (ctx *TestContext) IAmTheManagerOfTheGroup(groupName string) error {
	groupID := ctx.getOrCreateReferenceFor(groupName)

	return ctx.IAmTheManagerOfTheGroupWith(getParametersString(map[string]string{
		"id":                strconv.FormatInt(groupID, 10),
		"name":              groupName,
		"can_watch_members": "false",
	}))
}

// theGroupIsADescendantOfTheGroup sets a group as a descendant of another.
func (ctx *TestContext) theGroupIsADescendantOfTheGroup(descendant, parent string) error {
	// we add another group in between to increase the robustness of the tests.
	descendantID := ctx.getOrCreateReferenceFor(descendant)
	middleGroupID := rand.Int63()
	parentID := ctx.getOrCreateReferenceFor(parent)

	groupsIDs := []int64{descendantID, middleGroupID, parentID}
	for _, groupID := range groupsIDs {
		err := ctx.ThereIsAGroupWith(getParametersString(map[string]string{
			"id": strconv.FormatInt(groupID, 10),
		}))
		if err != nil {
			return err
		}
	}

	ctx.IsAMemberOfTheGroup(middleGroupID, parentID)
	ctx.IsAMemberOfTheGroup(descendantID, middleGroupID)

	return nil
}

// ICanWatchGroupWithID adds the permission for the user to watch a group.
func (ctx *TestContext) ICanWatchGroupWithID(groupID int64) error {
	return ctx.IAmTheManagerOfTheGroupWith(getParametersString(map[string]string{
		"id":                strconv.FormatInt(groupID, 10),
		"can_watch_members": "true",
	}))
}

// ICanWatchGroup adds the permission for the user to watch a group.
func (ctx *TestContext) ICanWatchGroup(groupName string) error {
	groupID := ctx.getOrCreateReferenceFor(groupName)

	return ctx.IAmTheManagerOfTheGroupWith(getParametersString(map[string]string{
		"id":                strconv.FormatInt(groupID, 10),
		"name":              groupName,
		"can_watch_members": "true",
	}))
}

// IsAMemberOfTheGroup Puts a group in a group.
func (ctx *TestContext) IsAMemberOfTheGroup(childGroupID, parentGroupID int64) {
	ctx.addGroupGroup(strconv.FormatInt(parentGroupID, 10), strconv.FormatInt(childGroupID, 10))
}

// IAmAMemberOfTheGroupWithID creates a group and add the user in it.
func (ctx *TestContext) IAmAMemberOfTheGroupWithID(groupID int64) error {
	err := ctx.ThereIsAGroupWith("id=" + strconv.FormatInt(groupID, 10))
	if err != nil {
		return err
	}

	ctx.IsAMemberOfTheGroup(ctx.userID, groupID)

	return nil
}

// UserIsAMemberOfTheGroup puts a user in a group.
func (ctx *TestContext) UserIsAMemberOfTheGroup(userName, groupName string) error {
	userID := ctx.getOrCreateReferenceFor(userName)

	err := ctx.ThereIsAUserWith(getParametersString(map[string]string{
		"group_id": strconv.FormatInt(userID, 10),
		"login":    userName,
	}))
	if err != nil {
		return err
	}

	groupID := ctx.getOrCreateReferenceFor(groupName)
	err = ctx.ThereIsAGroupWith(getParametersString(map[string]string{
		"id":   strconv.FormatInt(groupID, 10),
		"name": groupName,
	}))
	if err != nil {
		return err
	}

	ctx.IsAMemberOfTheGroup(userID, groupID)

	return nil
}

// IAmAMemberOfTheGroup puts a user in a group.
func (ctx *TestContext) IAmAMemberOfTheGroup(name string) error {
	return ctx.IAmAMemberOfTheGroupWithID(ctx.getOrCreateReferenceFor(name))
}

// ICanOnItemWithID gives the user a permission on an item.
func (ctx *TestContext) ICanOnItemWithID(watchType, watchValue string, itemID int64) error {
	ctx.addPermissionGenerated(strconv.FormatInt(ctx.userID, 10), strconv.FormatInt(itemID, 10), watchType, watchValue)

	return nil
}

// ICanViewOnItemWithID gives the user a "view" permission on an item.
func (ctx *TestContext) ICanViewOnItemWithID(watchValue string, itemID int64) error {
	return ctx.ICanOnItemWithID("view", watchValue, itemID)
}

// ICanWatchOnItemWithID gives the user a "watch" permission on an item.
func (ctx *TestContext) ICanWatchOnItemWithID(watchValue string, itemID int64) error {
	return ctx.ICanOnItemWithID("watch", watchValue, itemID)
}

// IHaveValidatedItemWithID states that user has validated an item.
func (ctx *TestContext) IHaveValidatedItemWithID(itemID int64) error {
	attemptID := rand.Int63()

	ctx.addAttempt(strconv.FormatInt(attemptID, 10), strconv.FormatInt(ctx.userID, 10))
	ctx.addResult(
		strconv.FormatInt(attemptID, 10),
		strconv.FormatInt(ctx.userID, 10),
		strconv.FormatInt(itemID, 10),
		time.Now(),
	)

	return nil
}

// ThereAreThreads create threads.
func (ctx *TestContext) ThereAreThreads(dataTable *messages.PickleStepArgument_PickleTable) error {
	if len(dataTable.Rows) > 1 {
		for i := 1; i < len(dataTable.Rows); i++ {
			err := ctx.ThereIsAThreadWith(getParametersString(ctx.getRowMap(i, dataTable)))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ThereIsAThreadWith creates a thread.
func (ctx *TestContext) ThereIsAThreadWith(parameters string) error {
	thread := ctx.getParametersMap(parameters)

	// add item
	if _, ok := thread["item_id"]; !ok {
		thread["item_id"] = strconv.FormatInt(rand.Int63(), 10)
	}
	ctx.addItem(thread["item_id"], "en", "Task")

	// add helper_group_id
	if _, ok := thread["helper_group_id"]; !ok {
		helperGroupID := rand.Int63()

		err := ctx.ThereIsAGroupWith(getParametersString(map[string]string{
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

	ctx.currentThreadKey = ctx.getThreadKey(thread["item_id"], thread["participant_id"])

	ctx.addThread(thread["item_id"], thread["participant_id"], thread["helper_group_id"], thread["status"], thread["message_count"])

	return nil
}

// ThereIsNoThreadWith states that a given thread doesn't exist.
func (ctx *TestContext) ThereIsNoThreadWith(parameters string) error {
	thread := ctx.getParametersMap(parameters)

	ctx.addItem(thread["item_id"], "en", "Task")

	return nil
}

// IAmPartOfTheHelperGroupOfTheThread states that user is a member of the helper group of a given thread.
func (ctx *TestContext) IAmPartOfTheHelperGroupOfTheThread() error {
	threadHelperGroupID := ctx.dbTables["threads"][ctx.currentThreadKey]["helper_group_id"]

	threadHelperGroupIDInt, _ := strconv.ParseInt(threadHelperGroupID.(string), 10, 64)
	ctx.IsAMemberOfTheGroup(ctx.userID, threadHelperGroupIDInt)

	return nil
}

// HasApprovedAccessPersonalInfoForGroup states that a user has approved the access of his personnal infos for a group.
func (ctx *TestContext) HasApprovedAccessPersonalInfoForGroup(userName, groupName string) error {
	err := ctx.UserIsAMemberOfTheGroup(userName, groupName)
	if err != nil {
		return err
	}

	groupID := ctx.getOrCreateReferenceFor(groupName)
	userID := ctx.getOrCreateReferenceFor(userName)

	ctx.addPersonalInfoViewApprovedFor(strconv.FormatInt(userID, 10), strconv.FormatInt(groupID, 10))

	return nil
}

// ICanRequestHelpToTheGroupWithIDOnTheItemWithID gives the user the permission to request help from a given group
// to a given item.
func (ctx *TestContext) ICanRequestHelpToTheGroupWithIDOnTheItemWithID(groupID, itemID int64) error {
	ctx.addPermissionGranted(
		strconv.FormatInt(ctx.userID, 10),
		strconv.FormatInt(ctx.userID, 10),
		strconv.FormatInt(itemID, 10),
		strconv.FormatInt(groupID, 10),
	)

	return nil
}
