Feature: Get navigation data (groupNavigationView) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name            | type  | is_public |
      | 1  | Team            | Team  | true      |
      | 2  | Managed By Team | Class | false     |
      | 41 | user            | User  | true      |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 41       | Jean-Michel | Blanquer  |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 2        | 1          |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | expires_at          |
      | 1               | 41             | 9999-12-31 23:59:59 |
    And the groups ancestors are computed

  Scenario: Invalid group_id given
    Given I am the user with id "41"
    When I send a GET request to "/groups/1_1/navigation"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Requires the group to be public/managed/descendant of managed/joined
    Given I am the user with id "41"
    When I send a GET request to "/groups/2/navigation"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Requires the group to exist
    Given I am the user with id "41"
    When I send a GET request to "/groups/404/navigation"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The group_id is a user
    Given I am the user with id "41"
    When I send a GET request to "/groups/41/navigation"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
