Feature: List attempts for current user and item_id - robustness
  Background:
    Given the database has the following users:
      | login | group_id | first_name | last_name |
      | jdoe  | 11       | John       | Doe       |
      | jane  | 12       | Jane       | Doe       |
    And the database has the following table 'groups':
      | id | type  |
      | 13 | Team  |
      | 14 | Class |
      | 15 | Team  |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 12             |
      | 14              | 12             |
      | 15              | 12             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 12                | 12             |
      | 13                | 12             |
      | 13                | 13             |
      | 14                | 12             |
      | 14                | 14             |
      | 15                | 12             |
      | 15                | 15             |
    And the database has the following table 'items':
      | id  | allows_multiple_attempts | default_language_tag |
      | 210 | 1                        | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 11       | 210     | content            |
      | 13       | 210     | info               |

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/1/attempts"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Wrong item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/abc/attempts"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Wrong sorting
    Given I am the user with id "11"
    When I send a GET request to "/items/210/attempts?sort=login"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "login""

  Scenario: User doesn't have access to the item
    Given I am the user with id "12"
    When I send a GET request to "/items/210/attempts?limit=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong as_team_id
    Given I am the user with id "12"
    When I send a GET request to "/items/210/attempts?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: Team doesn't have access to the item
    Given I am the user with id "12"
    When I send a GET request to "/items/210/attempts?as_team_id=13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User is not a member of the team
    Given I am the user with id "11"
    When I send a GET request to "/items/210/attempts?as_team_id=13"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: as_team_id is not a team
    Given I am the user with id "12"
    When I send a GET request to "/items/210/attempts?as_team_id=14"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"
    And the table "attempts" should stay unchanged
