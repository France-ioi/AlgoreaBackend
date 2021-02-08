Feature: Sign the current user out
  Scenario: The user logs out successfully
    Given the database has the following users:
      | group_id | login |
      | 2        | john  |
      | 3        | jane  |
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'sessions':
      | user_id | expires_at          | access_token              | use_cookie |
      | 2       | 2019-07-16 22:02:29 | someaccesstoken           | true       |
      | 2       | 2019-07-16 22:02:40 | anotheraccesstoken        | false      |
      | 3       | 2019-07-16 22:02:29 | accesstokenforjane        | false      |
      | 3       | 2019-07-16 22:02:31 | anotheraccesstokenforjane | false      |
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
    And the response header "Set-Cookie" should be "[NULL]"
    And the table "sessions" should be:
      | user_id | expires_at          | access_token              |
      | 3       | 2019-07-16 22:02:29 | accesstokenforjane        |
      | 3       | 2019-07-16 22:02:31 | anotheraccesstokenforjane |
    And the table "refresh_tokens" should be:
      | user_id | refresh_token       |
      | 3       | refreshtokenforjane |
    And the table "users" should stay unchanged

  Scenario Outline: The user logs out successfully with the session cookie provided
    Given the database has the following users:
      | group_id | login |
      | 2        | john  |
      | 3        | jane  |
    And the time now is "2019-07-16T22:02:28Z"
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'sessions':
      | user_id | expires_at          | access_token              | use_cookie | cookie_secure | cookie_same_site | cookie_domain | cookie_path |
      | 2       | 2019-07-16 22:02:29 | someaccesstoken           | true       | false         | false            | null          | null        |
      | 2       | 2019-07-16 22:02:40 | anotheraccesstoken        | false      | false         | false            | null          | null        |
      | 2       | 2019-07-16 22:02:40 | onemoreaccesstoken        | true       | true          | false            | a.127.0.0.1   | /api/       |
      | 2       | 2019-07-16 22:02:40 | thirdaccesstoken          | true       | false         | true             | 127.0.0.1     | /           |
      | 3       | 2019-07-16 22:02:29 | accesstokenforjane        | false      | false         | false            | null          | null        |
      | 3       | 2019-07-16 22:02:31 | anotheraccesstokenforjane | false      | false         | false            | null          | null        |
    And the database has the following table 'refresh_tokens':
      | user_id | refresh_token       |
      | 2       | somerefreshtoken    |
      | 3       | refreshtokenforjane |
    And the "Cookie" request header is "access_token=<access_token>"
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
      | user_id | expires_at          | access_token              |
      | 3       | 2019-07-16 22:02:29 | accesstokenforjane        |
      | 3       | 2019-07-16 22:02:31 | anotheraccesstokenforjane |
    And the table "refresh_tokens" should be:
      | user_id | refresh_token       |
      | 3       | refreshtokenforjane |
    And the table "users" should stay unchanged
    Examples:
      | access_token       | expected_cookie                                                                                                                  |
      | someaccesstoken    | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; SameSite=None                                         |
      | anotheraccesstoken | [NULL]                                                                                                                           |
      | onemoreaccesstoken | access_token=; Path=/api/; Domain=a.127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None |
      | thirdaccesstoken   | access_token=; Path=/; Domain=127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; SameSite=Strict             |
