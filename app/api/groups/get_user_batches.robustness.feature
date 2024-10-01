Feature: List user batches (userBatchesView) - robustness
  Background:
    Given the database has the following table "groups":
      | id | name   | grade | type  |
      | 13 | class  | -2    | Class |
      | 14 | class2 | -2    | Class |
      | 21 | user   | -2    | User  |
    And the database has the following table "users":
      | login | group_id |
      | owner | 21       |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
      | 14       | 21         | none        |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 13              | 21             |
    And the groups ancestors are computed

  Scenario: Invalid group_id given
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/1_1"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Invalid sorting rules given
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/13?sort=code"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "code""

  Scenario: A tie-breaker field is missing
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/13?sort=size&from.group_prefix=abc"
    Then the response code should be 400
    And the response error message should contain "All 'from' parameters (from.custom_prefix, from.group_prefix) or none of them must be present"
