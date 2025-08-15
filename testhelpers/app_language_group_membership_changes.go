//go:build !prod && !unit

package testhelpers

import (
	"fmt"

	"github.com/cucumber/godog"
)

// registerFeaturesForGroupMembershipChanges registers the Gherkin features related to group membership changes.
func (ctx *TestContext) registerFeaturesForGroupMembershipChanges(s *godog.ScenarioContext) {
	s.Step(`^there should be the following group membership changes:$`, ctx.ThereShouldBeTheFollowingGroupMembershipChanges)
}

// ThereShouldBeTheFollowingGroupMembershipChanges checks that rows are present in the group_membership_changes table of the database.
func (ctx *TestContext) ThereShouldBeTheFollowingGroupMembershipChanges(entries *godog.Table) error {
	for i := 1; i < len(entries.Rows); i++ {
		change := ctx.getRowMap(i, entries)

		resultCount := ctx.databaseCountRows("group_membership_changes", change)

		if resultCount == 0 {
			return fmt.Errorf("could not find the group membership change %+v", change)
		}
	}

	return nil
}
