Feature: Get managers of group_id - robustness
  Background:
    Given the database has the following users:
      | login | group_id |
      | owner | 21       |
      | user  | 11       |
    And the database has the following table "groups":
      | id |
      | 13 |
    And the database has the following table "group_managers":
      | group_id | manager_id |
      | 13       | 21         |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | expires_at          |
      | 13              | 11             | 2019-05-30 11:00:00 |
    And the groups ancestors are computed

  Scenario: User is not a manager of the group
    Given I am the user with id "11"
    When I send a GET request to "/groups/13/managers"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/managers"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Group id is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc/managers"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/managers?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

  Scenario: include_managers_of_ancestor_groups is invalid
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/managers?include_managers_of_ancestor_groups=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for include_managers_of_ancestor_groups (should have a boolean value (0 or 1))"
