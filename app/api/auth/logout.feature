Feature: Sign the current user out
  Scenario: The user logs out successfully
    Given the database has the following users:
      | group_id | login |
      | 2        | john  |
      | 3        | jane  |
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'sessions':
      | session_id | user_id | refresh_token       |
      | 1          | 2       | somerefreshtoken    |
      | 2          | 3       | refreshtokenforjane |
    And the database has the following table 'access_tokens':
      | session_id | token                     | expires_at          |
      | 1          | someaccesstoken           | 2019-07-16 22:02:29 |
      | 1          | anotheraccesstoken        | 2019-07-16 22:02:40 |
      | 2          | accesstokenforjane        | 2019-07-16 22:02:29 |
      | 2          | anotheraccesstokenforjane | 2019-07-16 22:02:31 |
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
    And the response header "Set-Cookie" should be "[NULL]"
    And the table "sessions" should be:
      | session_id | user_id | refresh_token       |
      | 2          | 3       | refreshtokenforjane |
    And the table "access_tokens" should be:
      | session_id | token                     | expires_at          |
      | 2          | accesstokenforjane        | 2019-07-16 22:02:29 |
      | 2          | anotheraccesstokenforjane | 2019-07-16 22:02:31 |
    And the table "users" should stay unchanged

  Scenario Outline: The user logs out successfully with the session cookie provided
    Given the database has the following users:
      | group_id | login |
      | 2        | john  |
      | 3        | jane  |
    And the time now is "2019-07-16T22:02:28Z"
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'sessions':
      | session_id | user_id | refresh_token       |
      | 1          | 2       | somerefreshtoken    |
      | 2          | 3       | refreshtokenforjane |
    And the database has the following table 'access_tokens':
      | session_id | token                     | expires_at          |
      | 1          | someaccesstoken           | 2019-07-16 22:02:29 |
      | 1          | anotheraccesstoken        | 2019-07-16 22:02:40 |
      | 1          | onemoreaccesstoken        | 2019-07-16 22:02:40 |
      | 1          | thirdaccesstoken          | 2019-07-16 22:02:40 |
      | 2          | accesstokenforjane        | 2019-07-16 22:02:29 |
      | 2          | anotheraccesstokenforjane | 2019-07-16 22:02:31 |
    And the "Cookie" request header is "access_token=<access_token_cookie>"
    When I send a POST request to "/auth/logout"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "success"
    }
    """
    And the response header "Set-Cookie" should be "<expected_cookie>"
    And the table "sessions" should be:
      | session_id | user_id | refresh_token       |
      | 2          | 3       | refreshtokenforjane |
    And the table "access_tokens" should be:
      | session_id | token                     | expires_at          |
      | 2          | accesstokenforjane        | 2019-07-16 22:02:29 |
      | 2          | anotheraccesstokenforjane | 2019-07-16 22:02:31 |
    And the table "users" should stay unchanged
    Examples:
      | access_token_cookie                    | expected_cookie                                                                                                                  |
      | 0!someaccesstoken!!                    | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; SameSite=None                                         |
      | 2!onemoreaccesstoken!a.127.0.0.1!/api/ | access_token=; Path=/api/; Domain=a.127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None |
      | 1!thirdaccesstoken!127.0.0.1!/         | access_token=; Path=/; Domain=127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; SameSite=Strict             |
