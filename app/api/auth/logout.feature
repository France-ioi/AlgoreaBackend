Feature: Sign the current user out
  Scenario: The user logs out successfully
    Given the database has the following table 'users':
      | id | login |
      | 2  | john  |
      | 3  | jane  |
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'sessions':
      | user_id | expires_at          | access_token              |
      | 2       | 2019-07-16 22:02:29 | someaccesstoken           |
      | 2       | 2019-07-16 22:02:40 | anotheraccesstoken        |
      | 3       | 2019-07-16 22:02:29 | accesstokenforjane        |
      | 3       | 2019-07-16 22:02:31 | anotheraccesstokenforjane |
    And the database has the following table 'refresh_tokens':
      | user_id | refresh_token       |
      | 2       | somerefreshtoken    |
      | 3       | refreshtokenforjane |
    And the "Authorization" request header is "Bearer someaccesstoken"
    When I send a POST request to "/auth/logout"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "success"
    }
    """
    And the table "sessions" should be:
      | user_id | expires_at          | access_token              |
      | 3       | 2019-07-16 22:02:29 | accesstokenforjane        |
      | 3       | 2019-07-16 22:02:31 | anotheraccesstokenforjane |
    And the table "refresh_tokens" should be:
      | user_id | refresh_token       |
      | 3       | refreshtokenforjane |
    And the table "users" should stay unchanged
