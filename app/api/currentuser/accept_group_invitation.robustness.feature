Feature: User accepts an invitation to join a group - robustness
  Background:
    Given the database has the following table 'groups':
      | id | type     | team_item_id | require_personal_info_access_approval |
      | 11 | Class    | null         | none                                  |
      | 13 | Friends  | null         | none                                  |
      | 14 | Team     | 1234         | none                                  |
      | 15 | Team     | 1234         | none                                  |
      | 16 | Team     | null         | view                                  |
      | 21 | UserSelf | null         | none                                  |
    And the database has the following table 'users':
      | group_id | login |
      | 21       | john  |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 13             |
      | 14                | 13             |
      | 14                | 14             |
      | 14                | 21             |
      | 21                | 13             |
      | 21                | 21             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 14              | 21             |
      | 21              | 13             |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 11       | 21        | join_request |
      | 13       | 21        | invitation   |
      | 15       | 21        | invitation   |
      | 16       | 21        | invitation   |

  Scenario: User tries to create a cycle in the group relations graph
    Given I am the user with id "21"
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
    Given I am the user with id "21"
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
    Given I am the user with id "21"
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
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-invitations/abc/accept"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails if the user doesn't exist
    Given I am the user with id "404"
    When I send a POST request to "/current-user/group-invitations/14/accept"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: User tries to accept an invitation to join a group that requires approvals which are not given
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-invitations/16/accept"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "error_text": "Missing required approvals",
      "data": {
        "missing_approvals": ["personal_info_view"]
      }
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
