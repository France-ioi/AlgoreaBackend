Feature: Check if a login id is the current user's login id
  Background:
    Given the database has the following users:
      | group_id | temp_user | login | login_id |
      | 2        | 0         | user  | 1234     |
      | 3        | 1         | jane  | null     |

  Scenario: Login ID matches
    Given I am the user with id "2"
    When I send a GET request to "/current-user/check-login-id?login_id=1234"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "login_id_matched": true
    }
    """

  Scenario: Login ID mismatches
    Given I am the user with id "2"
    When I send a GET request to "/current-user/check-login-id?login_id=12345"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "login_id_matched": false
    }
    """

  Scenario: Login ID is NULL
    Given I am the user with id "3"
    When I send a GET request to "/current-user/check-login-id?login_id=1234"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "login_id_matched": false
    }
    """
