//go:build !prod

package testhelpers

import (
	"fmt"

	"github.com/cucumber/godog"

	"github.com/cucumber/messages-go/v10"
)

// registerFeaturesForGroupMembershipChanges registers the Gherkin features related to group membership changes.
func (ctx *TestContext) registerFeaturesForGroupMembershipChanges(s *godog.Suite) {
	s.Step(`^there should be the following group membership changes:$`, ctx.ThereShouldBeTheFollowingGroupMembershipChanges)
}

// ThereShouldBeTheFollowingGroupMembershipChanges checks that rows are present in the group_membership_changes table in database.
func (ctx *TestContext) ThereShouldBeTheFollowingGroupMembershipChanges(entries *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(entries.Rows); i++ {
		change := ctx.getRowMap(i, entries)

		resultCount := ctx.databaseCountRows("group_membership_changes", change)

		if resultCount == 0 {
			return fmt.Errorf("could not find the group membership change %+v", change)
		}
	}

	return nil
}
