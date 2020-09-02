Feature: Check if the group code is valid - robustness
  Background:
    Given the database has the following table 'groups':
      | id | type  |
      | 21 | User  |
    And the database has the following table 'users':
      | login | group_id | temp_user |
      | john  | 21       | false     |

  Scenario: No code
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid"
    Then the response code should be 400
    And the response error message should contain "Missing code"
