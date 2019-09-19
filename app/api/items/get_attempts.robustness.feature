Feature: Get groups attempts for current user and item_id - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | group_self_id | group_owned_id | first_name | last_name |
      | 1  | jdoe  | 11            | 12             | John       | Doe       |

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/1/attempts"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Wrong item_id
    Given I am the user with id "1"
    When I send a GET request to "/items/abc/attempts"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Wrong sorting
    Given I am the user with id "1"
    When I send a GET request to "/items/123/attempts?sort=login"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "login""
