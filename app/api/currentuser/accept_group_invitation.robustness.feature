Feature: User accepts an invitation to join a group - robustness
  Background:
    Given the database has the following table 'groups':
      | id | type      | team_item_id |
      | 11 | Class     | null         |
      | 13 | Friends   | null         |
      | 14 | Team      | 1234         |
      | 15 | Team      | 1234         |
      | 21 | UserSelf  | null         |
      | 22 | UserAdmin | null         |
    And the database has the following table 'users':
      | group_id | owned_group_id | login |
      | 21       | 22             | john  |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 13             | 1       |
      | 14                | 13             | 0       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 21                | 13             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type               | type_changed_at     |
      | 1  | 11              | 21             | requestSent        | 2017-04-29 06:38:38 |
      | 2  | 13              | 21             | invitationSent     | 2017-03-29 06:38:38 |
      | 7  | 14              | 21             | invitationAccepted | 2017-02-21 06:38:38 |
      | 8  | 15              | 21             | invitationSent     | 2017-03-29 06:38:38 |
      | 10 | 21              | 13             | direct             | 2017-01-29 06:38:38 |

  Scenario: User tries to create a cycle in the group relations graph
    Given I am the user with group_id "21"
    When I send a POST request to "/current-user/group-invitations/13/accept"
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

  Scenario: User tries to accept an invitation that doesn't exist
    Given I am the user with group_id "21"
    When I send a POST request to "/current-user/group-invitations/11/accept"
    Then the response code should be 404
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Not Found",
      "error_text": "No such relation"
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User tries to accept an invitation to join a team while being a member of another team with the same team_item_id
    Given I am the user with group_id "21"
    When I send a POST request to "/current-user/group-invitations/15/accept"
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

  Scenario: Fails when the group id is wrong
    Given I am the user with group_id "21"
    When I send a POST request to "/current-user/group-invitations/abc/accept"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails if the user doesn't exist
    Given I am the user with group_id "404"
    When I send a POST request to "/current-user/group-invitations/14/accept"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

