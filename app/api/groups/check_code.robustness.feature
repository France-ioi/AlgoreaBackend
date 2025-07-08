Feature: Check if the group code is valid - robustness
  Background:
    Given the database has the following user:
      | group_id | login |
      | 21       | john  |

  Scenario: No code
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid"
    Then the response code should be 400
    And the response error message should contain "Missing code"
