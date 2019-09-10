Feature: User sends a request to join a group - robustness
  Background:
    Given the database has the following table 'users':
      | ID | idGroupSelf | idGroupOwned | sLogin |
      | 1  | 21          | 22           | john   |
      | 2  | null        | null         | guest  |
    And the database has the following table 'groups':
      | ID | bFreeAccess | sType     | idTeamItem |
      | 11 | 1           | Class     | null       |
      | 13 | 1           | Friends   | null       |
      | 14 | 1           | Team      | 1234       |
      | 15 | 0           | Club      | null       |
      | 16 | 1           | Team      | 1234       |
      | 17 | 0           | Team      | 1234       |
      | 21 | 0           | UserSelf  | null       |
      | 22 | 0           | UserAdmin | null       |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 15              | 15           | 1       |
      | 16              | 16           | 1       |
      | 16              | 21           | 0       |
      | 17              | 17           | 1       |
      | 21              | 13           | 0       |
      | 21              | 21           | 1       |
      | 22              | 17           | 0       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          |
      | 1  | 11            | 21           | invitationSent     | 2017-04-29T06:38:38Z |
      | 7  | 14            | 21           | requestSent        | 2017-02-21T06:38:38Z |
      | 8  | 16            | 21           | invitationAccepted | 2017-02-21T06:38:38Z |
      | 9  | 21            | 13           | direct             | 2017-01-29T06:38:38Z |
      | 10 | 22            | 17           | direct             | 2017-01-29T06:38:38Z |

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
    And the table "groups_ancestors" should stay unchanged

  Scenario: User tries to send a request while a conflicting relation exists
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
    And the table "groups_ancestors" should stay unchanged

  Scenario: User tries to send a request to join a team while being a member of another team with the same idTeamItem
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/14"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "error_text": "You are already on a team for this item"
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Team owner tries to send a request to join a team while being a member of another team with the same idTeamItem
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/17"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "error_text": "You are already on a team for this item"
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the group ID is wrong
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

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
    And the table "groups_ancestors" should stay unchanged

  Scenario: Can't send request to a group having bFreeAccess=0
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/15"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
