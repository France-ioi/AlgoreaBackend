Feature: Join a group using a password (groupsJoinByPassword) - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned |
      | 1  | john   | 21          | 22           |
      | 2  | nobody | null        | null         |
    And the database has the following table 'groups':
      | ID | sType     | sPassword  | sPasswordEnd         | sPasswordTimer | bFreeAccess |
      | 11 | Team      | 3456789abc | 2017-04-29T06:38:38Z | null           | true        |
      | 12 | Team      | abc3456789 | null                 | null           | true        |
      | 14 | Team      | cba9876543 | null                 | null           | true        |
      | 15 | Team      | 75987654ab | null                 | null           | false       |
      | 16 | Class     | dcef123492 | null                 | null           | false       |
      | 21 | UserSelf  | null       | null                 | null           | false       |
      | 22 | UserAdmin | null       | null                 | null           | false       |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 12              | 12           | 1       |
      | 14              | 14           | 1       |
      | 14              | 21           | 0       |
      | 15              | 15           | 1       |
      | 16              | 16           | 1       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          |
      | 1  | 11            | 21           | invitationSent     | 2017-04-29T06:38:38Z |
      | 7  | 14            | 21           | invitationAccepted | 2017-02-21T06:38:38Z |

  Scenario: No password
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-password"
    Then the response code should be 400
    And the response error message should contain "Missing password"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User's self group is NULL
    Given I am the user with ID "2"
    When I send a POST request to "/current-user/group-memberships/by-password?password=cba9876543"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join with a wrong password
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-password?password=abcdef"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with ID = 1 tried to join a group using a wrong/expired password
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join with an expired password
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-password?password=3456789abc"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with ID = 1 tried to join a group using a wrong/expired password
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join a group that is not a team
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-password?password=dcef123492"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with ID = 1 tried to join a group using a wrong/expired password
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join a closed team
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-password?password=75987654ab"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And logs should contain:
      """
      A user with ID = 1 tried to join a group using a wrong/expired password
      """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Join an already joined group
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-password?password=cba9876543"
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
