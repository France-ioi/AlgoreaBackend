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

		var conditions string
		var values []interface{}
		for key, value := range change {
			if conditions != "" {
				conditions += " AND "
			}

			conditions += key + " = ? "

			if value[0] == ReferencePrefix {
				values = append(values, ctx.getReference(value))
			} else {
				values = append(values, value)
			}
		}

		query := "SELECT COUNT(*) as count FROM `group_membership_changes` WHERE " + conditions

		var resultCount int
		err := db.QueryRow(query, values...).
			Scan(&resultCount)
		if err != nil {
			return err
		}

		if resultCount == 0 {
			return fmt.Errorf("could not find the group membership change %+v", change)
		}
	}

	return nil
}
