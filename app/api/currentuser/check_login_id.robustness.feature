Feature: Check if a login id is the current user's login id - robustness
  Background:
    Given the database has the following users:
      | group_id | temp_user | login | login_id |
      | 2        | 0         | user  | 1234     |
      | 3        | 1         | jane  | null     |

  Scenario: Login ID is invalid
    Given I am the user with id "2"
    When I send a GET request to "/current-user/check-login-id?login_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for login_id (should be int64)"
