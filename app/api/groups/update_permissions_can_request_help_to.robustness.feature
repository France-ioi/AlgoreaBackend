# Those scenario cannot, for now, be merged with those in update_permissions.feature
#
# Reason: The scenario in this file are defined with new Gherkin features which allows higher-level definitions.
#         Those features require the propagation of permissions to run.
# Problem: the permissions defined in update_permissions.feature contain inconsistent data.
#          It means that if we move the definitions of the table permissions_generated into the equivalent in permissions_granted,
#          and then we run the propagation of permissions, we get a different result than
#          the permissions currently defined in permissions_generated, and many tests then fail.
#          If those permissions definitions get fixed, then this file can be merged with them.
Feature: Change item access rights for a group - can_request_help_to
  Background:
    Given allUsersGroup is defined as the group @AllUsers
    And there are the following groups:
      | group        | parent       | members  |
      | @AllUsers    |              |          |
      | @School      |              | @Teacher |
      | @Class       | @ClassParent |          |
      | @HelperGroup |              |          |
    And @Teacher is a manager of the group @ClassParent and can grant group access
    And there are the following tasks:
      | item  |
      | @Item |
    And there are the following item permissions:
      | item  | group  | can_view |
      | @Item | @Class | info     |

  Scenario: Should be an exception when can_request_help_to is not an int64
    Given I am @Teacher
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": "aaa"
        }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid input data"

  Scenario: Should be an exception when can_request_help_to_all_users is not a boolean
    Given I am @Teacher
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": {
            "is_all_users_group": 1
          }
        }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid input data"

  Scenario: Should be an exception when trying to set can_request_help_to and can_request_help_to_all_users at the same time
    Given I am @Teacher
    # @HelperGroup is visible by @Teacher
    And the group @Teacher is a descendant of the group @HelperGroup via @HelperGroupChild
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": {
            "id": "@HelperGroup",
            "is_all_users_group": true
          }
        }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "can_request_help_to": ["cannot set can_request_help_to id and is_all_users_group at the same time"]
        }
      }
      """

  Scenario Outline: Should be access denied when the user doesn't have can_grant_view>=content but group is visible by both the current user and the receiver
    Given I am @Teacher
    # @HelperGroup is visible by @Teacher
    And the group @Teacher is a descendant of the group @HelperGroup via @HelperGroupChild
    # @HelperGroup is visible by @Class
    And the group @Class is a descendant of the group @HelperGroup via @HelperGroupAnotherChild
    And there is a group @OldHelperGroup
    And there are the following item permissions:
      | item  | group    | can_grant_view | can_view | can_request_help_to |
      | @Item | @Teacher | enter          | info     |                     |
      | @Item | @Class   | none           | info     | @OldHelperGroup     |
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": {
            "id": <helper_group>
          }
        }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "can_request_help_to": ["the current user doesn't have the right to update can_request_help_to"]
        }
      }
      """
    Examples:
      | helper_group |
      | @HelperGroup |
      | @AllUsers    |
      | null         |

  Scenario: Should be access denied when trying to set can_request_help_to to a group not visible by the giver (current-user)
    Given I am @Teacher
    # This is the only case for @HelperGroup to be visible by @Class and not @Teacher. Details in comment in update_permissions.go.
    And @Class is a manager of the group @HelperGroupParent and can watch its members
    And @HelperGroup is a child of the group @HelperGroupParent
    And there are the following item permissions:
      | item  | group    | can_grant_view |
      | @Item | @Teacher | content        |
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": {
            "id": "@HelperGroup"
          }
        }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "can_request_help_to": ["can_request_help_to is not visible either by the current-user or the groupID"]
        }
      }
      """

  Scenario: Should be access denied when trying to set can_request_help_to to a group no visible by the receiver
    Given I am @Teacher
    # @HelperGroup is visible by @Teacher
    And the group @Teacher is a descendant of the group @HelperGroup via @HelperGroupChild
    And there are the following item permissions:
      | item  | group    | can_grant_view |
      | @Item | @Teacher | content        |
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": {
            "id": "@HelperGroup"
          }
        }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "can_request_help_to": ["can_request_help_to is not visible either by the current-user or the groupID"]
        }
      }
      """
