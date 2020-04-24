Feature: Join a group using a code (groupsJoinByCode) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | type  | code       | code_expires_at     | code_lifetime | is_public | require_watch_approval | frozen_membership |
      | 11 | Team  | 3456789abc | 2017-04-29 06:38:38 | null          | true      | 0                      | false             |
      | 12 | Team  | abc3456789 | null                | null          | true      | 1                      | false             |
      | 14 | Team  | cba9876543 | null                | null          | true      | 0                      | false             |
      | 15 | Team  | 75987654ab | null                | null          | false     | 0                      | false             |
      | 16 | Class | dcef123492 | null                | null          | false     | 0                      | false             |
      | 17 | Team  | 5987654abc | null                | null          | true      | 0                      | false             |
      | 18 | Team  | 87654abcde | null                | null          | true      | 0                      | true              |
      | 21 | User  | null       | null                | null          | false     | 0                      | false             |
      | 22 | User  | null       | null                | null          | false     | 0                      | false             |
    And the database has the following table 'users':
      | login | group_id | temp_user |
      | john  | 21       | false     |
      | tmp   | 22       | true      |
    And the database has the following table 'items':
      | id   | default_language_tag |
      | 1234 | fr                   |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 14              | 21             |
    And the groups ancestors are computed
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type       |
      | 11       | 21        | invitation |
    And the database has the following table 'attempts':
      | participant_id | id | root_item_id |
      | 14             | 1  | 1234         |
      | 17             | 2  | 1234         |

  Scenario: No code
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code"
    Then the response code should be 400
    And the response error message should contain "Missing code"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join with a wrong code
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=abcdef"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with group_id = 21 tried to join a group using a wrong/expired code
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join with an expired code
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=3456789abc"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with group_id = 21 tried to join a group using a wrong/expired code
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join a group that is not a team
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=dcef123492"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with group_id = 21 tried to join a group using a wrong/expired code
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: The user is temporary
    Given I am the user with id "22"
    When I send a POST request to "/current-user/group-memberships/by-code?code=cba9876543"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join a closed team
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=75987654ab"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with group_id = 21 tried to join a group using a wrong/expired code
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join an already joined group
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=cba9876543"
    Then the response code should be 422
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Unprocessable Entity",
        "error_text": "A conflicting relation exists"
      }
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join a team while being a member of another team participating in same contests
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=5987654abc"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "error_text": "Team's participations are in conflict with the user's participations"
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Cannot join if required approvals are missing
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=abc3456789"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "data": {"missing_approvals":["watch"]},
      "error_text": "Missing required approvals"
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Cannot join if the group membership is frozen
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=87654abcde"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "error_text": "Group membership is frozen"
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
