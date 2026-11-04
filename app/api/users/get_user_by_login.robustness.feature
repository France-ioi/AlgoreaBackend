Feature: Get user info by login - robustness
  Background:
    Given the database has the following users:
      | group_id | temp_user | login | default_language |
      | 2        | 0         | user  | en               |
      | 3        | 1         | jane  | fr               |
      | 4        | 0         | john  | fr               |

  Scenario: User not found
    Given I am the user with id "2"
    When I send a GET request to "/users/by-login/404"
    Then the response code should be 404
    And the response error message should contain "No such user"
