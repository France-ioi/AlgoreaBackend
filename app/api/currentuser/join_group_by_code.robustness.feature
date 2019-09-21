Feature: Join a group using a code (groupsJoinByCode) - robustness
  Background:
    Given the database has the following table 'users':
      | id | login  | self_group_id | owned_group_id |
      | 1  | john   | 21            | 22             |
      | 2  | nobody | null          | null           |
    And the database has the following table 'groups':
      | id | type      | code       | code_end            | code_timer | free_access | team_item_id |
      | 11 | Team      | 3456789abc | 2017-04-29 06:38:38 | null       | true        | null         |
      | 12 | Team      | abc3456789 | null                | null       | true        | null         |
      | 14 | Team      | cba9876543 | null                | null       | true        | 1234         |
      | 15 | Team      | 75987654ab | null                | null       | false       | null         |
      | 16 | Class     | dcef123492 | null                | null       | false       | null         |
      | 17 | Team      | 5987654abc | null                | null       | true        | 1234         |
      | 21 | UserSelf  | null       | null                | null       | false       | null         |
      | 22 | UserAdmin | null       | null                | null       | false       | null         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 15                | 15             | 1       |
      | 16                | 16             | 1       |
      | 17                | 17             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type               | status_date         |
      | 1  | 11              | 21             | invitationSent     | 2017-04-29 06:38:38 |
      | 7  | 14              | 21             | invitationAccepted | 2017-02-21 06:38:38 |

  Scenario: No code
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code"
    Then the response code should be 400
    And the response error message should contain "Missing code"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User's self group is NULL
    Given I am the user with id "2"
    When I send a POST request to "/current-user/group-memberships/by-code?code=cba9876543"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join with a wrong code
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=abcdef"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with id = 1 tried to join a group using a wrong/expired code
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join with an expired code
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=3456789abc"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with id = 1 tried to join a group using a wrong/expired code
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join a group that is not a team
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=dcef123492"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with id = 1 tried to join a group using a wrong/expired code
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join a closed team
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=75987654ab"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with id = 1 tried to join a group using a wrong/expired code
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join an already joined group
    Given I am the user with id "1"
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

  Scenario: Join a team while being a member of another team with the same team_item_id
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=5987654abc"
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
