Feature: Get permissions granted to group - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name          | type  |
      | 21 | owner         | User  |
      | 23 | user          | User  |
      | 25 | some class    | Class |
      | 26 | another class | Class |
      | 27 | third class   | Class |
      | 31 | admin         | User  |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | user  | 23       | John        | Doe       |
      | admin | 31       | Allie       | Grater    |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_grant_group_access |
      | 23       | 21         | 1                      |
      | 25       | 21         | 0                      |
      | 25       | 31         | 1                      |
      | 26       | 21         | 1                      |
      | 26       | 31         | 1                      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 25              | 23             |
      | 25              | 31             |
      | 26              | 23             |
    And the groups ancestors are computed

  Scenario: Invalid group_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/111111111111111111111111111111111/granted_permissions"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Invalid descendants
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/granted_permissions?descendants=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for descendants (should have a boolean value (0 or 1))"

  Scenario: The user is not a manager of the group_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/27/granted_permissions"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is a manager of the group_id, but he doesn't have 'can_grant_group_access' permission on its ancestors
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/granted_permissions"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is a manager of the group_id with 'can_grant_group_access' permission, but the group_id group is a user
    Given I am the user with id "21"
    When I send a GET request to "/groups/23/granted_permissions"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/26/granted_permissions?sort=name"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "name""
