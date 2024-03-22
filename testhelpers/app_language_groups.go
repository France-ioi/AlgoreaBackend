//go:build !prod

package testhelpers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/cucumber/messages-go/v10"
)

// registerFeaturesForGroups registers the Gherkin features related to groups.
func (ctx *TestContext) registerFeaturesForGroups(s *godog.Suite) {
	s.Step(`^there are the following groups:$`, ctx.ThereAreTheFollowingGroups)
	s.Step(`^there is a group with "([^"]*)"$`, ctx.ThereIsAGroupWith)
	s.Step(`^there is a group (@\w+)$`, ctx.ThereIsAGroup)
	s.Step(`^I am a member of the group (@\w+)$`, ctx.IAmAMemberOfTheGroup)
	s.Step(`^I am a member of the group with id "([^"]*)"$`, ctx.IAmAMemberOfTheGroupWithID)
	s.Step(`^(@\w+) is a member of the group (@\w+)$`, ctx.UserIsAMemberOfTheGroup)
	s.Step(
		`^(@\w+) is a member of the group (@\w+) who has approved access to his personal info$`,
		ctx.UserIsAMemberOfTheGroupWhoHasApprovedAccessToHisPersonalInfo,
	)
	s.Step(`allUsersGroup is defined as the group (@\w+)$`, ctx.AllUsersGroupIsDefinedAsTheGroup)

	s.Step(`^the field "([^"]*)" of the group (@\w+) should be "([^"]*)"$`, ctx.TheFieldOfTheGroupShouldBe)
	s.Step(`^(@\w+) should not be a member of the group (@\w+)$`, ctx.UserShouldNotBeAMemberOfTheGroup)
	s.Step(`^(@\w+) should be a member of the group (@\w+)$`, ctx.UserShouldBeAMemberOfTheGroup)

	s.Step(`^the group (@\w+) is a descendant of the group (@\w+)$`, ctx.theGroupIsADescendantOfTheGroup)
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

// ThereAreTheFollowingGroups defines groups.
func (ctx *TestContext) ThereAreTheFollowingGroups(groups *messages.PickleStepArgument_PickleTable) error {
	for i := 1; i < len(groups.Rows); i++ {
		group := ctx.getRowMap(i, groups)

		groupParameters := map[string]string{
			"id":   group["group"],
			"name": group["group"],
		}
		if _, ok := group["require_personal_info_access_approval"]; ok {
			groupParameters["require_personal_info_access_approval"] = group["require_personal_info_access_approval"]
		}
		if _, ok := group["require_lock_membership_approval_until"]; ok {
			groupParameters["require_lock_membership_approval_until"] = group["require_lock_membership_approval_until"]
		}
		if _, ok := group["require_watch_approval"]; ok {
			groupParameters["require_watch_approval"] = group["require_watch_approval"]
		}

		err := ctx.ThereIsAGroupWith(getParameterString(groupParameters))
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

// ThereIsAGroup creates a new group.
func (ctx *TestContext) ThereIsAGroup(group string) error {
	return ctx.ThereIsAGroupWith(getParameterString(map[string]string{
		"id":   group,
		"name": group,
	}))
}

// IAmAMemberOfTheGroup puts a user in a group.
func (ctx *TestContext) IAmAMemberOfTheGroup(name string) error {
	return ctx.IAmAMemberOfTheGroupWithID(name)
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
