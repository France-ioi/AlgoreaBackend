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
      | group                    | parent                   | members  |
      | @AllUsers                |                          |          |
      | @School                  |                          | @Teacher |
      | @ClassParentParentParent |                          |          |
      | @ClassParentParent       | @ClassParentParentParent |          |
      | @ClassParent             | @ClassParentParent       |          |
      | @Class                   | @ClassParent             |          |
      | @OtherSourceGroup        |                          |          |
    And the group @Teacher is a manager of the group @ClassAnotherParent and can grant group access
    And the group @Class is a child of the group @ClassAnotherParent
    And there are the following items:
      | item                        | type    |
      | @ChapterParent              | Chapter |
      | @ChapterParentNoPropagation | Chapter |
      | @Chapter                    | Chapter |
      | @Item                       | Task    |
    And there are the following item relations:
      | item     | parent                      | request_help_propagation |
      | @Item    | @Chapter                    | true                     |
      | @Chapter | @ChapterParent              | true                     |
      | @Chapter | @ChapterParentNoPropagation | false                    |
    And there are the following item permissions:
      | item  | group    | is_owner | can_request_help_to |
      | @Item | @Teacher | true     |                     |

  Scenario: Should return helper group when set and visible by the current user
    Given I am @Teacher
    And there is a group @HelperGroup
    # @HelperGroup is visible by @Teacher
    And the group @Teacher is a descendant of the group @HelperGroup via @HelperGroupChild
    And there are the following item permissions:
      | item  | group  | is_owner | can_request_help_to |
      | @Item | @Class | false    | @HelperGroup        |
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted.can_request_help_to in JSON should be:
      """
        {
          "id": "@HelperGroup",
          "name": "Group HelperGroup",
          "is_all_users_group": false
        }
      """

  Scenario: Should return helper group without the name when set and not visible by the current user
    Given I am @Teacher
    And there is a group @HelperGroup
    And there are the following item permissions:
      | item  | group  | is_owner | can_request_help_to |
      | @Item | @Class | false    | @HelperGroup        |
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted.can_request_help_to in JSON should be:
      """
        {
          "id": "@HelperGroup",
          "is_all_users_group": false
        }
      """

  Scenario: Should return helper group as "AllUsers" group when set to its value
    Given I am @Teacher
    And there are the following item permissions:
      | item  | group  | is_owner | can_request_help_to |
      | @Item | @Class | false    | @AllUsers           |
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted.can_request_help_to in JSON should be:
      """
        {
          "id": "@AllUsers",
          "name": "AllUsers",
          "is_all_users_group": true
        }
      """

  Scenario: Should return can_request_help_to arrays when permissions with specific origins are defined
    Given I am @Teacher
    And there are the following item permissions:
      | item  | group                    | source_group      | origin           | is_owner | can_request_help_to               | comment                                           |
      | @Item | @Class                   | @Class            | self             | false    | @HelperGroupSelf1                 |                                                   |
      | @Item | @ClassParentParentParent | @Class            | self             | false    | @HelperGroupSelf1                 | shouldn't contain duplicate of previous one       |
      | @Item | @ClassParent             | @Class            | self             | false    | @AllUsers                         |                                                   |
      | @Item | @ClassParentParent       | @Class            | self             | false    |                                   | check we don't get empty groups                   |
      | @Item | @ClassParentParent       | @OtherSourceGroup | self             | false    | @HelperOtherSourceGroup1          | other source group                                |
      | @Item | @Class                   | @Class            | group_membership | false    | @HelperGroupGroupMembership1      | in granted and computed, but not group_membership |
      | @Item | @ClassParent             | @Class            | group_membership | false    | @HelperGroupGroupMembership2      | group_membership but not granted                  |
      | @Item | @ClassParentParent       | @Class            | group_membership | false    | @AllUsers                         | group_membership but not granted                  |
      | @Item | @ClassParentParent       | @OtherSourceGroup | group_membership | false    | @HelperOtherSourceGroup2          | other source group                                |
      | @Item | @ClassParent             | @Class            | item_unlocking   | false    | @HelperGroupItemUnlocking         |                                                   |
      | @Item | @Class                   | @Class            | item_unlocking   | false    | @AllUsers                         |                                                   |
      | @Item | @ClassParentParent       | @Class            | item_unlocking   | false    | @HelperGroupNotVisible            | not visible                                       |
      | @Item | @ClassParentParent       | @OtherSourceGroup | item_unlocking   | false    | @HelperOtherSourceGroup3          | other source group                                |
      | @Item | @Class                   | @Class            | other            | false    | @AllUsers                         |                                                   |
      | @Item | @ClassParentParent       | @Class            | other            | false    | @HelperGroupOther1                |                                                   |
      | @Item | @ClassParent             | @Class            | other            | false    | @HelperGroupOther2                |                                                   |
      | @Item | @ClassParent             | @OtherSourceGroup | other            | false    | @HelperOtherSourceGroup4          | other source group                                |
      | @Item | @ClassParentParent       | @OtherSourceGroup | other            | false    | @HelperOtherSourceGroupNotVisible | other source group, not visible                   |
  # The following lines are to make the groups visible by @Teacher
  And the group @Teacher is a descendant of the group @HelperGroupSelf1 via @HelperGroupSelf1Child
  And the group @Teacher is a descendant of the group @HelperGroupGroupMembership1 via @HelperGroupGroupMembership1Child
  And the group @Teacher is a descendant of the group @HelperGroupGroupMembership2 via @HelperGroupGroupMembership2Child
  And the group @Teacher is a descendant of the group @HelperGroupItemUnlocking via @HelperGroupItemUnlockingChild
  And the group @Teacher is a descendant of the group @HelperGroupOther1 via @HelperGroupOther1Child
  And the group @Teacher is a descendant of the group @HelperGroupOther2 via @HelperGroupOther2Child
  And the group @Teacher is a descendant of the group @HelperOtherSourceGroup1 via @HelperOtherSourceGroup1Child
  And the group @Teacher is a descendant of the group @HelperOtherSourceGroup2 via @HelperOtherSourceGroup2Child
  And the group @Teacher is a descendant of the group @HelperOtherSourceGroup3 via @HelperOtherSourceGroup3Child
  And the group @Teacher is a descendant of the group @HelperOtherSourceGroup4 via @HelperOtherSourceGroup4Child
  When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
  Then the response code should be 200
  And the response at $.granted_via_self.can_request_help_to[*] should be:
    | id                       | name                          | is_all_users_group |
    | @AllUsers                | AllUsers                      | true               |
    | @HelperGroupSelf1        | Group HelperGroupSelf1        | false              |
    | @HelperOtherSourceGroup1 | Group HelperOtherSourceGroup1 | false              |
  And the response at $.granted.can_request_help_to in JSON should be:
    """
      {
        "id": "@HelperGroupGroupMembership1",
        "name": "Group HelperGroupGroupMembership1",
        "is_all_users_group": false
      }
    """
  And the response at $.granted_via_group_membership.can_request_help_to[*] should be:
    | id                           | name                              | is_all_users_group |
    | @HelperGroupGroupMembership2 | Group HelperGroupGroupMembership2 | false              |
    | @AllUsers                    | AllUsers                          | true               |
    | @HelperOtherSourceGroup2     | Group HelperOtherSourceGroup2     | false              |
  And the response at $.granted_via_item_unlocking.can_request_help_to[*] should be:
    | id                        | name                           | is_all_users_group |
    | @HelperGroupItemUnlocking | Group HelperGroupItemUnlocking | false              |
    | @AllUsers                 | AllUsers                       | true               |
    | @HelperGroupNotVisible    | <undefined>                    | false              |
    | @HelperOtherSourceGroup3  | Group HelperOtherSourceGroup3  | false              |
  And the response at $.granted_via_other.can_request_help_to[*] should be:
    | id                                | name                          | is_all_users_group |
    | @AllUsers                         | AllUsers                      | true               |
    | @HelperGroupOther1                | Group HelperGroupOther1       | false              |
    | @HelperGroupOther2                | Group HelperGroupOther2       | false              |
    | @HelperOtherSourceGroup4          | Group HelperOtherSourceGroup4 | false              |
    | @HelperOtherSourceGroupNotVisible | <undefined>                   | false              |
  And the response at $.computed.can_request_help_to[*] should be:
    | id                                | name                              | is_all_users_group |
    | @AllUsers                         | AllUsers                          | true               |
    | @HelperGroupSelf1                 | Group HelperGroupSelf1            | false              |
    | @HelperGroupGroupMembership1      | Group HelperGroupGroupMembership1 | false              |
    | @HelperGroupGroupMembership2      | Group HelperGroupGroupMembership2 | false              |
    | @HelperGroupItemUnlocking         | Group HelperGroupItemUnlocking    | false              |
    | @HelperGroupNotVisible            | <undefined>                       | false              |
    | @HelperGroupOther1                | Group HelperGroupOther1           | false              |
    | @HelperGroupOther2                | Group HelperGroupOther2           | false              |
    | @HelperOtherSourceGroup1          | Group HelperOtherSourceGroup1     | false              |
    | @HelperOtherSourceGroup2          | Group HelperOtherSourceGroup2     | false              |
    | @HelperOtherSourceGroup3          | Group HelperOtherSourceGroup3     | false              |
    | @HelperOtherSourceGroup4          | Group HelperOtherSourceGroup4     | false              |
    | @HelperOtherSourceGroupNotVisible | <undefined>                       | false              |

  Scenario: Should return can_request_help_to from parent items only into computed and only when it propagates
    Given I am @Teacher
    And there are the following item permissions:
      | item                        | group                    | source_group      | origin           | is_owner | can_request_help_to                      | comment            |
      | @Chapter                    | @Class                   | @Class            | self             | false    | @HelperGroupSelf1                        |                    |
      | @ChapterParent              | @ClassParentParentParent | @Class            | self             | false    | @HelperGroupSelfNotVisible               | without name       |
      | @ChapterParentNoPropagation | @ClassParent             | @Class            | self             | false    | @HelperGroupSelfNoPropagation            | shouldn't appear   |
      | @Chapter                    | @Class                   | @Class            | group_membership | false    | @HelperGroupGroupMembership1             |                    |
      | @ChapterParent              | @ClassParent             | @Class            | group_membership | false    | @AllUsers                                |                    |
      | @ChapterParentNoPropagation | @ClassParentParent       | @Class            | group_membership | false    | @HelperGroupGroupMembershipNoPropagation | shouldn't appear   |
      | @Chapter                    | @ClassParent             | @Class            | item_unlocking   | false    | @HelperGroupItemUnlocking1               |                    |
      | @ChapterParent              | @Class                   | @Class            | item_unlocking   | false    | @HelperGroupItemUnlocking2               |                    |
      | @ChapterParentNoPropagation | @ClassParentParent       | @Class            | item_unlocking   | false    | @HelperGroupItemUnlockingNoPropagation   | shouldn't appear   |
      | @Chapter                    | @Class                   | @Class            | other            | false    | @HelperGroupOther1                       |                    |
      | @ChapterParent              | @ClassParentParent       | @Class            | other            | false    | @HelperGroupOther2                       |                    |
      | @ChapterParentNoPropagation | @ClassParent             | @OtherSourceGroup | other            | false    | @HelperGroupOtherNoPropagation           | other source group |
  # The following lines are to make the groups visible by @Teacher
    And the group @Teacher is a descendant of the group @HelperGroupSelf1 via @HelperGroupSelf1Child
    And the group @Teacher is a descendant of the group @HelperGroupGroupMembership1 via @HelperGroupGroupMembership1Child
    And the group @Teacher is a descendant of the group @HelperGroupGroupMembership2 via @HelperGroupGroupMembership2Child
    And the group @Teacher is a descendant of the group @HelperGroupItemUnlocking1 via @HelperGroupItemUnlocking1Child
    And the group @Teacher is a descendant of the group @HelperGroupItemUnlocking2 via @HelperGroupItemUnlocking2Child
    And the group @Teacher is a descendant of the group @HelperGroupOther1 via @HelperGroupOther1Child
    And the group @Teacher is a descendant of the group @HelperGroupOther2 via @HelperGroupOther2Child
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted_via_self.can_request_help_to should be "[]"
    And the response at $.granted.can_request_help_to should be "<null>"
    And the response at $.granted_via_group_membership.can_request_help_to should be "[]"
    And the response at $.granted_via_item_unlocking.can_request_help_to should be "[]"
    And the response at $.granted_via_other.can_request_help_to should be "[]"
    And the response at $.computed.can_request_help_to[*] should be:
      | id                           | name                              | is_all_users_group |
      | @HelperGroupSelf1            | Group HelperGroupSelf1            | false              |
      | @HelperGroupSelfNotVisible   | <undefined>                       | false              |
      | @HelperGroupGroupMembership1 | Group HelperGroupGroupMembership1 | false              |
      | @AllUsers                    | AllUsers                          | true               |
      | @HelperGroupItemUnlocking1   | Group HelperGroupItemUnlocking1   | false              |
      | @HelperGroupItemUnlocking2   | Group HelperGroupItemUnlocking2   | false              |
      | @HelperGroupOther1           | Group HelperGroupOther1           | false              |
      | @HelperGroupOther2           | Group HelperGroupOther2           | false              |
