# Those scenario cannot, for now, be merged with those in get_permissions.feature
#
# Reason: The scenario in this file are defined with new Gherkin features which allows higher-level definitions.
#         Those features require the propagation of permissions to run.
# Problem: the permissions defined in get_permissions.feature contain inconsistent data.
#          It means that if we move the definitions of the table permissions_generated into the equivalent in permissions_granted,
#          and then we run the propagation of permissions, we get a different result than
#          the permissions currently defined in permissions_generated, and many tests then fail.
#          If those permissions definitions get fixed, then this file can be merged with them.
Feature: Get permissions can_request_help_to for a group
  Background:
    Given allUsersGroup is defined as the group @AllUsers
    And there are the following groups:
      | group              | parent             | members  |
      | @AllUsers          |                    |          |
      | @School            |                    | @Teacher |
      | @ClassParentParent |                    |          |
      | @ClassParent       | @ClassParentParent |          |
      | @Class             | @ClassParent       |          |
    And @Teacher is a manager of the group @Class and can grant group access
    And there are the following tasks:
      | item  |
      | @Item |
    And there are the following item permissions:
      | item  | group    | is_owner | can_request_help_to |
      | @Item | @Teacher | true     |                     |

  Scenario: Should return helper group when set and visible by the current user
    Given I am @Teacher
    And there is a group @HelperGroup
    # @HelperGroup is visible by @Teacher
    And the group @Teacher is a descendant of the group @HelperGroup
    And there are the following item permissions:
      | item  | group  | is_owner | can_request_help_to |
      | @Item | @Class | false    | @HelperGroup        |
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted.can_request_help_to should be:
      | id           | name              | is_all_users_group |
      | @HelperGroup | Group HelperGroup | false              |

  Scenario: Should return helper group without the name when set and not visible by the current user
    Given I am @Teacher
    And there is a group @HelperGroup
    And there are the following item permissions:
      | item  | group  | is_owner | can_request_help_to |
      | @Item | @Class | false    | @HelperGroup        |
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted.can_request_help_to should be:
      | id           | name | is_all_users_group |
      | @HelperGroup |      | false              |

  Scenario: Should return helper group as "AllUsers" group when set to its value
    Given I am @Teacher
    And there are the following item permissions:
      | item  | group  | is_owner | can_request_help_to |
      | @Item | @Class | false    | @AllUsers           |
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted.can_request_help_to should be:
      | id        | name     | is_all_users_group |
      | @AllUsers | AllUsers | true               |

  Scenario: Should return can_request_help_to arrays when permissions with specific origins are defined
    Given I am @Teacher
    And there are the following item permissions:
      | item  | group              | source_group | origin           | is_owner | can_request_help_to          | comment                          |
      | @Item | @Class             | @Class       | self             | false    | @HelperGroupSelf1            |                                  |
      | @Item | @ClassParent       | @Class       | self             | false    | @AllUsers                    |                                  |
      | @Item | @ClassParentParent | @Class       | self             | false    |                              | check we don't get empty groups  |
      | @Item | @Class             | @Class       | group_membership | false    | @HelperGroupGroupMembership1 | granted                          |
      | @Item | @ClassParent       | @Class       | group_membership | false    | @HelperGroupGroupMembership2 | group_membership but not granted |
      | @Item | @ClassParentParent | @Class       | group_membership | false    | @AllUsers                    | group_membership but not granted |
      | @Item | @ClassParent       | @Class       | item_unlocking   | false    | @HelperGroupItemUnlocking    |                                  |
      | @Item | @Class             | @Class       | item_unlocking   | false    | @AllUsers                    |                                  |
      | @Item | @ClassParentParent | @Class       | item_unlocking   | false    | @HelperGroupNotVisible       | not visible                      |
      | @Item | @Class             | @Class       | other            | false    | @AllUsers                    |                                  |
      | @Item | @ClassParentParent | @Class       | other            | false    | @HelperGroupOther1           |                                  |
      | @Item | @ClassParent       | @Class       | other            | false    | @HelperGroupOther2           |                                  |
  # The following lines are to make the groups visible by @Teacher
  And the group @Teacher is a descendant of the group @HelperGroupSelf1
  And the group @Teacher is a descendant of the group @HelperGroupGroupMembership1
  And the group @Teacher is a descendant of the group @HelperGroupGroupMembership2
  And the group @Teacher is a descendant of the group @HelperGroupItemUnlocking
  And the group @Teacher is a descendant of the group @HelperGroupOther1
  And the group @Teacher is a descendant of the group @HelperGroupOther2
  When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
  Then the response code should be 200
  And the response at $.granted_via_self.can_request_help_to[*] should be:
    | id                | name                   | is_all_users_group |
    | @AllUsers         | AllUsers               | true               |
    | @HelperGroupSelf1 | Group HelperGroupSelf1 | false              |
  And the response at $.granted.can_request_help_to should be:
    | id                           | name                              |
    | @HelperGroupGroupMembership1 | Group HelperGroupGroupMembership1 |
  And the response at $.granted_via_group_membership.can_request_help_to[*] should be:
    | id                           | name                              | is_all_users_group |
    | @HelperGroupGroupMembership2 | Group HelperGroupGroupMembership2 | false              |
    | @AllUsers                    | AllUsers                          | true               |
  And the response at $.granted_via_item_unlocking.can_request_help_to[*] should be:
    | id                        | name                           | is_all_users_group |
    | @HelperGroupItemUnlocking | Group HelperGroupItemUnlocking | false              |
    | @AllUsers                 | AllUsers                       | true               |
    | @HelperGroupNotVisible    |                                | false              |
  And the response at $.granted_via_other.can_request_help_to[*] should be:
    | id                 | name                    | is_all_users_group |
    | @AllUsers          | AllUsers                | true               |
    | @HelperGroupOther1 | Group HelperGroupOther1 | false              |
    | @HelperGroupOther2 | Group HelperGroupOther2 | false              |
