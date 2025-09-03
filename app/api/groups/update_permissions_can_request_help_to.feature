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
    Given the application config is:
      """
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: @AllUsers
      """
    And there are the following groups:
      | group           | parent       | members                  |
      | @AllUsers       |              | @NonTempUsers,@TempUsers |
      | @School         |              | @Teacher                 |
      | @Class          | @ClassParent |                          |
      | @OldHelperGroup |              |                          |
      | @NewHelperGroup |              |                          |
    And the group @Teacher is a manager of the group @ClassParent and can grant group access
    And the group @Teacher is a child of the group @NonTempUsers
    And there are the following tasks:
      | item  |
      | @Item |
    And there are the following item permissions:
      | item  | group    | can_view | can_grant_view |
      | @Item | @Teacher | info     | content        |

  Scenario Outline: Should update can_request_help_to to the desired value when rights are appropriate
    Given I am @Teacher
    # @OldHelperGroup is visible by @Teacher
    And the group @Teacher is a descendant of the group @OldHelperGroup via @OldHelperGroupChild1
    # @OldHelperGroup is visible by @Class
    And the group @Class is a descendant of the group @OldHelperGroup via @OldHelperGroupChild2
    # @NewHelperGroup is visible by @Teacher
    And the group @Teacher is a descendant of the group @NewHelperGroup via @NewHelperGroupChild1
    # @NewHelperGroup is visible by @Class
    And the group @Class is a descendant of the group @NewHelperGroup via @NewHelperGroupChild2
    And there are the following item permissions:
      | item  | group  | can_view | can_request_help_to           |
      | @Item | @Class | info     | <initial_can_request_help_to> |
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": {
            "id": <changed_can_request_help_to_request>
          }
        }
      """
    Then the response code should be 200
    Then the response should be "updated"
    And the table "permissions_granted" at group_id "@Class" should be:
      | group_id | item_id | source_group_id | can_request_help_to              |
      | @Class   | @Item   | @Class          | <changed_can_request_help_to_db> |
    Examples:
      | initial_can_request_help_to | changed_can_request_help_to_request | changed_can_request_help_to_db |
      |                             | "@NewHelperGroup"                   | @NewHelperGroup                |
      |                             | null                                | null                           |
      | @OldHelperGroup             | "@NewHelperGroup"                   | @NewHelperGroup                |
      | @OldHelperGroup             | null                                | null                           |
      | @OldHelperGroup             | "@OldHelperGroup"                   | @OldHelperGroup                |

  Scenario: Should update can_request_help_to to AllUsers group when specified
    Given I am @Teacher
    And there are the following item permissions:
      | item  | group  | can_view | can_request_help_to |
      | @Item | @Class | info     |                     |
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": {
            "is_all_users_group": true
          }
        }
      """
    Then the response code should be 200
    Then the response should be "updated"
    And the table "permissions_granted" at group_id "@Class" should be:
      | group_id | item_id | source_group_id | can_request_help_to |
      | @Class   | @Item   | @Class          | @AllUsers           |

  Scenario: Should work when trying to set can_request_help_to to a group not visible by the giver (current-user) if it was already set at the same value previously
    Given I am @Teacher
    # This is the only case for @HelperGroup to be visible by @Class and not @Teacher. Details in comment in update_permissions.go.
    And the group @Class is a manager of the group @HelperGroupParent and can watch for submissions from the group and its descendants
    And the group @HelperGroup is a child of the group @HelperGroupParent
    And there are the following item permissions:
      | item  | group    | can_view | can_grant_view | can_request_help_to |
      | @Item | @Teacher |          | content        |                     |
      | @Item | @Class   | info     |                | @HelperGroup        |
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
    """
      {
        "can_request_help_to": {
          "id": "@HelperGroup"
        }
      }
    """
    Then the response code should be 200
    Then the response should be "updated"

  Scenario: Should work when trying to set can_request_help_to to a group no visible by the receiver if it was already set at the same value previously
    Given I am @Teacher
    # @HelperGroup is visible by @Teacher
    And the group @Teacher is a descendant of the group @HelperGroup via @HelperGroupChild
    And there are the following item permissions:
      | item  | group    | can_view | can_grant_view | can_request_help_to |
      | @Item | @Teacher |          | content        |                     |
      | @Item | @Class   | info     |                | @HelperGroup        |
    When I send a PUT request to "/groups/@Class/permissions/@Class/@Item" with the following body:
      """
        {
          "can_request_help_to": {
            "id": "@HelperGroup"
          }
        }
      """
    Then the response code should be 200
    Then the response should be "updated"
