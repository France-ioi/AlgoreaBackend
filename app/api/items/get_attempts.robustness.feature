Feature: Get groups attempts for current user and item_id - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName | sLastName |
      | 1  | jdoe   | 11          | 12           | John       | Doe       |

  Scenario: Wrong item_id
    Given I am the user with ID "1"
    When I send a GET request to "/items/abc/attempts"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Wrong sorting
    Given I am the user with ID "1"
    When I send a GET request to "/items/123/attempts?sort=login"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "login""
