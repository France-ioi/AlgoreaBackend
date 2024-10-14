Feature: List groups managed by the current user - robustness
  Background:
    Given the database has the following table "groups":
      | id | name          | type  | description |
      | 21 | owner         | User  | null        |
    And the database has the following table "users":
      | login | group_id |
      | owner | 21       |

  Scenario: Wrong sort
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?sort=description"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "description""

  Scenario: Wrong from
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?from.type=Class"
    Then the response code should be 400
    And the response error message should contain "Unallowed paging parameters (from.type)"
