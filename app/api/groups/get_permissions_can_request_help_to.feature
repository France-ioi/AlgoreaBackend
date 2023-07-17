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
      | group     | parent | members  |
      | @AllUsers |        |          |
      | @School   |        | @Teacher |
      | @Class    |        |          |
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
      | id           | name              |
      | @HelperGroup | Group HelperGroup |

  Scenario: Should not return helper group when set and not visible by the current user
    Given I am @Teacher
    And there is a group @HelperGroup
    And there are the following item permissions:
      | item  | group  | is_owner | can_request_help_to |
      | @Item | @Class | false    | @HelperGroup        |
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted.can_request_help_to should be "null"

  Scenario: Should return helper group as "AllUsers" group when set to its value
    Given I am @Teacher
    And there are the following item permissions:
      | item  | group  | is_owner | can_request_help_to |
      | @Item | @Class | false    | @AllUsers           |
    When I send a GET request to "/groups/@Class/permissions/@Class/@Item"
    Then the response code should be 200
    And the response at $.granted.can_request_help_to should be:
      | is_all_users_group |
      | true               |
