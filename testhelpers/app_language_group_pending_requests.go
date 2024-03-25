//go:build !prod

package testhelpers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/godog"

	"github.com/cucumber/messages-go/v10"
)

// registerFeaturesForGroupMembershipChanges registers the Gherkin features related to group membership changes.
func (ctx *TestContext) registerFeaturesForGroupPendingRequests(s *godog.Suite) {
	s.Step(`^there are the following group pending requests:$`, ctx.ThereAreTheFollowingGroupPendingRequests)

	s.Step(
		`^there should be no group pending requests for the group (@\w+) with the type "([^"]*)"$`,
		ctx.ThereShouldBeNoGroupPendingRequestsForTheGroupWithTheType,
	)
	s.Step(`^there should be the following group pending requests:$`, ctx.ThereShouldBeTheFollowingGroupPendingRequests)
}

// getGroupPendingRequestPrimaryKey returns the primary key of a group pending request.
func (ctx *TestContext) getGroupPendingRequestPrimaryKey(groupID, memberID int64) string {
	return strconv.FormatInt(groupID, 10) + "," + strconv.FormatInt(memberID, 10)
}

// addGroup adds a group in database.
func (ctx *TestContext) addGroupPendingRequest(group, member, requestType string) {
	groupID := ctx.getReference(group)
	memberID := ctx.getReference(member)

	primaryKey := ctx.getGroupPendingRequestPrimaryKey(groupID, memberID)

	if !ctx.isInDatabase("group_pending_requests", primaryKey) {
		ctx.addInDatabase("group_pending_requests", primaryKey, map[string]interface{}{
			"group_id":  groupID,
			"member_id": memberID,
			"type":      requestType,
			// All the other fields are set to default values.
			"at": time.Now(),
		})
	}
}

// ThereAreTheFollowingGroupPendingRequests adds the group pending requests in the database.
func (ctx *TestContext) ThereAreTheFollowingGroupPendingRequests(entries *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(entries.Rows); i++ {
		change := ctx.getRowMap(i, entries)

		ctx.addGroupPendingRequest(change["group"], change["member"], change["type"])
	}

	return nil
}

// ThereShouldBeNoGroupPendingRequestsForTheGroupWithTheType checks that no rows are present in the group_pending_requests table.
func (ctx *TestContext) ThereShouldBeNoGroupPendingRequestsForTheGroupWithTheType(group, requestType string) error {
	groupID := ctx.getReference(group)

	query := "SELECT COUNT(*) as count FROM `group_pending_requests` WHERE `group_id` = ? AND `type` = ?"

	var resultCount int
	err := db.QueryRow(query, groupID, requestType).
		Scan(&resultCount)
	if err != nil {
		return err
	}

	if resultCount != 0 {
		return fmt.Errorf("found group pending requests for the group %s with the type %s", group, requestType)
	}

	return nil
}

// ThereShouldBeTheFollowingGroupPendingRequests checks that rows are present in the group_pending_requests table in database.
func (ctx *TestContext) ThereShouldBeTheFollowingGroupPendingRequests(entries *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(entries.Rows); i++ {
		change := ctx.getRowMap(i, entries)

		var conditions string
		var values []interface{}
		for key, value := range change {
			if conditions != "" {
				conditions += " AND "
			}

			conditions += "`" + key + "`" + " = ? "

			if value[0] == ReferencePrefix {
				values = append(values, ctx.getReference(value))
			} else {
				values = append(values, value)
			}
		}

		query := "SELECT COUNT(*) as count FROM `group_pending_requests` WHERE " + conditions

		var resultCount int
		err := db.QueryRow(query, values...).
			Scan(&resultCount)
		if err != nil {
			return err
		}

		if resultCount == 0 {
			return fmt.Errorf("could not find the group pending request %+v", change)
		}
	}

	return nil
}
