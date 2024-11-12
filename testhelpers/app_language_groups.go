//go:build !prod

package testhelpers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
)

// registerFeaturesForGroups registers the Gherkin features related to groups.
func (ctx *TestContext) registerFeaturesForGroups(s *godog.ScenarioContext) {
	s.Step(`^there are the following groups:$`, ctx.ThereAreTheFollowingGroups)
	s.Step(`^there is a group (@\w+)$`, ctx.ThereIsAGroup)
	s.Step(`^there is a team (@\w+)$`, ctx.ThereIsATeam)
	s.Step(`^I am a member of the group (@\w+)$`, ctx.IAmAMemberOfTheGroup)
	s.Step(`^I am a member of the group with id "([^"]*)"$`, ctx.IAmAMemberOfTheGroupWithID)
	s.Step(`^(@\w+) is a member of the group (@\w+)$`, ctx.UserIsAMemberOfTheGroup)
	s.Step(`^the group (@\w+) is a child of the group (@\w+)$`, ctx.GroupIsAChildOfTheGroup)
	s.Step(
		`^(@\w+) is a member of the group (@\w+) who has approved access to his personal info$`,
		ctx.UserIsAMemberOfTheGroupWhoHasApprovedAccessToHisPersonalInfo,
	)
	s.Step(`allUsersGroup is defined as the group (@\w+)$`, ctx.AllUsersGroupIsDefinedAsTheGroup)

	s.Step(`^the field "([^"]*)" of the group (@\w+) should be "([^"]*)"$`, ctx.TheFieldOfTheGroupShouldBe)
	s.Step(`^(@\w+) should not be a member of the group (@\w+)$`, ctx.UserShouldNotBeAMemberOfTheGroup)
	s.Step(`^(@\w+) should be a member of the group (@\w+)$`, ctx.UserShouldBeAMemberOfTheGroup)

	s.Step(`^the group (@\w+) is a descendant of the group (@\w+) via (@\w+)$`, ctx.theGroupIsADescendantOfTheGroup)
}

// getGroupPrimaryKey returns the primary key of a group.
func (ctx *TestContext) getGroupPrimaryKey(groupID int64) map[string]string {
	return map[string]string{"id": strconv.FormatInt(groupID, 10)}
}

// addGroup adds a group to the database.
func (ctx *TestContext) addGroup(group, groupType string) {
	groupID := ctx.getIDOfReference(group)
	primaryKey := ctx.getGroupPrimaryKey(groupID)

	if !ctx.isInDatabase("groups", primaryKey) {
		ctx.needPopulateDatabase = true
		err := ctx.DBHasTable("groups", &godog.Table{
			Rows: []*messages.PickleTableRow{
				{Cells: []*messages.PickleTableCell{
					{Value: "id"},
					{Value: "name"},
					{Value: "type"},
					{Value: "require_personal_info_access_approval"},
					{Value: "require_lock_membership_approval_until"},
					{Value: "require_watch_approval"},
				}},
				{Cells: []*messages.PickleTableCell{
					{Value: strconv.FormatInt(groupID, 10)},
					{Value: "Group " + referenceToName(group)},
					{Value: groupType},
					{Value: "none"},
					{Value: "null"},
					{Value: "false"},
				}},
			},
		})
		if err != nil {
			panic(err)
		}
	}
}

// setGroupFieldInDatabase sets a specific field of a group in the database.
func (ctx *TestContext) setGroupFieldInDatabase(primaryKey map[string]string, field, value string) {
	ctx.setDBTableRowColumnValue("groups", primaryKey, field, value)
}

// ThereAreTheFollowingGroups defines groups.
func (ctx *TestContext) ThereAreTheFollowingGroups(groups *godog.Table) error {
	for i := 1; i < len(groups.Rows); i++ {
		group := ctx.getRowMap(i, groups)
		groupID := ctx.getIDOfReference(group["group"])

		err := ctx.ThereIsAGroup(group["group"])
		mustNotBeError(err)

		if _, ok := group["require_personal_info_access_approval"]; ok {
			ctx.setGroupFieldInDatabase(
				ctx.getGroupPrimaryKey(groupID),
				"require_personal_info_access_approval",
				group["require_personal_info_access_approval"],
			)
		}
		if _, ok := group["require_lock_membership_approval_until"]; ok {
			ctx.setGroupFieldInDatabase(
				ctx.getGroupPrimaryKey(groupID),
				"require_lock_membership_approval_until",
				group["require_lock_membership_approval_until"],
			)
		}
		if _, ok := group["require_watch_approval"]; ok {
			ctx.setGroupFieldInDatabase(
				ctx.getGroupPrimaryKey(groupID),
				"require_watch_approval",
				group["require_watch_approval"],
			)
		}

		if _, ok := group["parent"]; ok {
			err = ctx.GroupIsAChildOfTheGroup(group["group"], group["parent"])
			mustNotBeError(err)
		}

		if _, ok := group["members"]; ok {
			members := strings.Split(group["members"], ",")

			for _, member := range members {
				err = ctx.ThereIsAUser(member)
				mustNotBeError(err)

				err = ctx.GroupIsAChildOfTheGroup(member, group["group"])
				mustNotBeError(err)
			}
		}
	}

	return nil
}

// ThereIsAGroup creates a new group (type=Class).
func (ctx *TestContext) ThereIsAGroup(group string) (err error) {
	defer recoverPanics(&err)
	ctx.addGroup(group, "Class")
	return nil
}

// ThereIsATeam creates a new team.
func (ctx *TestContext) ThereIsATeam(group string) (err error) {
	defer recoverPanics(&err)
	ctx.addGroup(group, "Team")
	return nil
}

// IAmAMemberOfTheGroup puts a user in a group.
func (ctx *TestContext) IAmAMemberOfTheGroup(name string) error {
	return ctx.IAmAMemberOfTheGroupWithID(name)
}

// IAmAMemberOfTheGroupWithID creates a group and add the user in it.
func (ctx *TestContext) IAmAMemberOfTheGroupWithID(group string) error {
	err := ctx.ThereIsAGroup(group)
	if err != nil {
		return err
	}

	ctx.GroupIsAMemberOfTheGroup(
		ctx.user,
		group,
	)

	return nil
}

// GroupIsAChildOfTheGroup puts a group as a child of another group.
func (ctx *TestContext) GroupIsAChildOfTheGroup(childGroup, parentGroup string) error {
	err := ctx.ThereIsAGroup(childGroup)
	if err != nil {
		return err
	}

	err = ctx.ThereIsAGroup(parentGroup)
	if err != nil {
		return err
	}

	ctx.GroupIsAMemberOfTheGroup(childGroup, parentGroup)

	return nil
}

// UserIsAMemberOfTheGroup puts a user in a group.
func (ctx *TestContext) UserIsAMemberOfTheGroup(user, group string) error {
	err := ctx.ThereIsAUser(user)
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
func (ctx *TestContext) AllUsersGroupIsDefinedAsTheGroup(group string) (err error) {
	defer recoverPanics(&err)

	ctx.addGroup(group, "Base")

	groupPrimaryKey := ctx.getGroupPrimaryKey(ctx.getIDOfReference(group))
	ctx.setGroupFieldInDatabase(groupPrimaryKey, "name", "AllUsers")

	err = ctx.TheApplicationConfigIs(&godog.DocString{
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
	resultCount := ctx.databaseCountRows("groups", map[string]string{
		"id":  group,
		field: value,
	})

	if resultCount != 1 {
		return fmt.Errorf("expected the group %s have %s=%s", group, field, value)
	}

	return nil
}

// UserShouldNotBeAMemberOfTheGroup checks that the user is not a member of the group.
func (ctx *TestContext) UserShouldNotBeAMemberOfTheGroup(user, group string) error {
	resultCount := ctx.databaseCountRows("groups_groups", map[string]string{
		"parent_group_id": group,
		"child_group_id":  user,
	})

	if resultCount != 0 {
		return fmt.Errorf("expected the user %s not to be a member of the group %s", user, group)
	}

	return nil
}

// UserShouldBeAMemberOfTheGroup checks that the user is a member of the group.
func (ctx *TestContext) UserShouldBeAMemberOfTheGroup(user, group string) error {
	resultCount := ctx.databaseCountRows("groups_groups", map[string]string{
		"parent_group_id": group,
		"child_group_id":  user,
	})

	if resultCount != 1 {
		return fmt.Errorf("expected the user %s to be a member of the group %s", user, group)
	}

	return nil
}

// theGroupIsADescendantOfTheGroup sets a group as a descendant of another via a third group.
func (ctx *TestContext) theGroupIsADescendantOfTheGroup(descendant, parent, middle string) error {
	groups := []string{descendant, middle, parent}
	for _, group := range groups {
		err := ctx.ThereIsAGroup(group)
		if err != nil {
			return err
		}
	}

	ctx.GroupIsAMemberOfTheGroup(middle, parent)
	ctx.GroupIsAMemberOfTheGroup(descendant, middle)

	return nil
}
