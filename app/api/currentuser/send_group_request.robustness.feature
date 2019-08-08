Feature: User sends a request to join a group - robustness
  Background:
    Given the database has the following table 'users':
      | ID | idGroupSelf | idGroupOwned | sLogin |
      | 1  | 21          | 22           | john   |
      | 2  | null        | null         | guest  |
    And the database has the following table 'groups':
      | ID | bFreeAccess |
      | 11 | 1           |
      | 13 | 1           |
      | 14 | 1           |
      | 15 | 0           |
      | 21 | 0           |
      | 22 | 0           |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType          | sStatusDate          |
      | 1  | 11            | 21           | invitationSent | 2017-04-29T06:38:38Z |
      | 7  | 14            | 21           | requestSent    | 2017-02-21T06:38:38Z |
      | 8  | 21            | 13           | direct         | 2017-01-29T06:38:38Z |

  Scenario: User tries to create a cycle in the group relations graph
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/13"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "error_text": "Cycles in the group relations graph are not allowed"
    }
    """
    And the table "groups_groups" should stay unchanged

  Scenario: User tries to send another conflicting relation exists
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/11"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "error_text": "A conflicting relation exists"
    }
    """
    And the table "groups_groups" should stay unchanged

  Scenario: Fails when the group ID is wrong
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should stay unchanged

  Scenario: Fails if the user doesn't exist
    Given I am the user with ID "4"
    When I send a POST request to "/current-user/group-requests/14"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Fails when the user's idGroupSelf is NULL
    Given I am the user with ID "2"
    When I send a POST request to "/current-user/group-requests/14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged

  Scenario: Can't send request to a group having bFreeAccess=0
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/15"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
