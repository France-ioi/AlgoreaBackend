Feature: Get user info - robustness
  Background:
    Given the database has the following users:
      | group_id | temp_user | login | first_name | last_name | default_language | free_text | web_site   |
      | 2        | 0         | user  | John       | Doe       | en               | Some text | mysite.com |
      | 3        | 1         | jane  | null       | null      | fr               | null      | null       |
      | 4        | 0         | john  | null       | null      | fr               | null      | null       |

  Scenario: Invalid user_id
    Given I am the user with id "2"
    When I send a GET request to "/users/123456789012345678901234567890"
    Then the response code should be 400
    And the response error message should contain "Wrong value for user_id (should be int64)"

  Scenario: No authentication
    Given I send a GET request to "/users/123456789012345678901234567890"
    Then the response code should be 401
    And the response error message should contain "No access token provided"

  Scenario: User not found
    Given I am the user with id "2"
    When I send a GET request to "/users/404"
    Then the response code should be 404
    And the response error message should contain "No such user"
