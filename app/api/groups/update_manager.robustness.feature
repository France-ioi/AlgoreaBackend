Feature: Update the group manager's permissions (groupManagerEdit) - robustness

  Background:
    Given the database has the following table "groups":
      | id | name    | type    |
      | 1  | Group   | Class   |
      | 2  | Team    | Team    |
      | 3  | Friends | Friends |
      | 21 | owner   | User    |
      | 22 | john    | User    |
    And the database has the following table "users":
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | john  | 22       | John        | Doe       |
    And the database has the following table "group_managers":
      | manager_id | group_id | can_manage            |
      | 21         | 1        | memberships_and_group |
      | 21         | 3        | memberships           |

  Scenario: group_id is wrong
    Given I am the user with id "21"
    When I send a PUT request to "/groups/abc/managers/22" with the following body:
      """
      {}
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "group_managers" should stay unchanged

  Scenario: manager_id is wrong
    Given I am the user with id "21"
    When I send a PUT request to "/groups/2/managers/abc" with the following body:
      """
      {}
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for manager_id (should be int64)"
    And the table "group_managers" should stay unchanged

  Scenario: Wrong JSON
    Given I am the user with id "21"
    When I send a PUT request to "/groups/2/managers/22" with the following body:
      """
      {
      """
    Then the response code should be 400
    And the response error message should contain "Unexpected EOF"
    And the table "group_managers" should stay unchanged

  Scenario: manager_id doesn't exist
    Given I am the user with id "21"
    When I send a PUT request to "/groups/2/managers/404" with the following body:
      """
      {}
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "group_managers" should stay unchanged

  Scenario: The user doesn't have enough permissions on the group
    Given I am the user with id "21"
    When I send a PUT request to "/groups/3/managers/22" with the following body:
      """
      {}
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "group_managers" should stay unchanged

  Scenario: There group_id-manager_id pair doesn't exist in group_managers
    Given I am the user with id "21"
    When I send a PUT request to "/groups/1/managers/22" with the following body:
      """
      {}
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "group_managers" should stay unchanged
