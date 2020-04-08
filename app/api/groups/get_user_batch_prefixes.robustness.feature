Feature: List user-batch prefixes (userBatchPrefixesView) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name   | grade | type  |
      | 13 | class  | -2    | Class |
      | 14 | class2 | -2    | Class |
      | 21 | user   | -2    | User  |
    And the database has the following table 'users':
      | login | group_id |
      | owner | 21       |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
      | 14       | 21         | none        |
      | 21       | 21         | memberships |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 21             |
    And the groups ancestors are computed

  Scenario: Invalid group_id given
    Given I am the user with id "21"
    When I send a GET request to "/groups/1_1/user-batch-prefixes"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: User is not a manager of the group_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/14/user-batch-prefixes"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User is has enough permissions to manage the group, but the group is a user
    Given I am the user with id "21"
    When I send a GET request to "/groups/21/user-batch-prefixes"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Invalid sorting rules given
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/user-batch-prefixes?sort=code"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "code""
