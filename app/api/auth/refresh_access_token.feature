Feature: Create a new access token
  Background:
    Given the database has the following table 'groups':
      | id | name        | type |
      | 12 | tmp-1234567 | User |
      | 13 | jane        | User |
      | 14 | john        | User |
    And the database has the following table 'users':
      | group_id | login       | temp_user |
      | 12       | tmp-1234567 | true      |
      | 13       | jane        | false     |
      | 14       | john        | false     |
    And the time now is "2020-01-01T02:00:00Z"
    And the DB time now is "2020-01-01 02:00:00"
    And the database has the following table 'sessions':
      | session_id | user_id | refresh_token             |
      | 1          | 12      |                           |
      | 2          | 13      | jane_current_refreshtoken |
      | 3          | 14      | john_current_refreshtoken |
    And the database has the following table 'access_tokens':
      | session_id | issued_at           | expires_at          | token              |
      | 1          | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | tmp_old_token      |
      | 1          | 2020-01-01 01:00:12 | 2020-01-01 03:00:12 | tmp_current_token  |
      | 2          | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | jane_old_token     |
      | 2          | 2020-01-01 01:50:00 | 2020-01-01 03:50:00 | jane_current_token |
      | 3          | 2020-01-01 01:50:00 | 2020-01-01 03:50:00 | john_current_token |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      """

  Scenario Outline: Request a new access token for a temporary user
    Given the generated auth key is "tmp_new_token"
    And the "Authorization" request header is "Bearer tmp_current_token"
    And the "Cookie" request header is "<current_cookie>"
    When I send a POST request to "/auth/token<query>"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {<token_in_data> "expires_in": 7200}
      }
      """
    And the response header "Set-Cookie" should be "<expected_cookie>"
    And logs should contain:
      """
      Refreshed a session token expiring in 7200 seconds for a temporary user with group_id = 12
      """
    And the table "sessions" should be:
      | session_id | user_id | refresh_token             |
      | 1          | 12      |                           |
      | 2          | 13      | jane_current_refreshtoken |
      | 3          | 14      | john_current_refreshtoken |
    And the table "access_tokens" should be:
      | session_id | issued_at           | expires_at          | token              |
      | 1          | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | tmp_old_token      |
      | 1          | 2020-01-01 01:00:12 | 2020-01-01 03:00:12 | tmp_current_token  |
      | 1          | 2020-01-01 02:00:00 | 2020-01-01 04:00:00 | tmp_new_token      |
      | 2          | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | jane_old_token     |
      | 2          | 2020-01-01 01:50:00 | 2020-01-01 03:50:00 | jane_current_token |
      | 3          | 2020-01-01 01:50:00 | 2020-01-01 03:50:00 | john_current_token |
    Examples:
      | query                            | current_cookie        | token_in_data                   | expected_cookie                                                                                                                                          |
    |                                  | [Header not defined]  | "access_token":"tmp_new_token", | [Header not defined]                                                                                                                                     |
    | ?use_cookie=1&cookie_secure=1    | [Header not defined]  |                                 | access_token=2!tmp_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 04:00:00 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None |
    | ?use_cookie=1&cookie_same_site=1 | [Header not defined]  |                                 | access_token=1!tmp_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 04:00:00 GMT; Max-Age=7200; HttpOnly; SameSite=Strict       |
    | ?use_cookie=0                    | access_token=0!1234!! | "access_token":"tmp_new_token", | access_token=; Expires=Wed, 01 Jan 2020 01:43:20 GMT; Max-Age=0; HttpOnly; SameSite=None                                                                 |

  Scenario Outline: Request a new access token for a normal user
    Given the login module "token" endpoint for refresh token "jane_current_refreshtoken" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622400,
        "access_token":"jane_new_token",
        "refresh_token":"jane_new_refreshtoken"
      }
      """
    And the "Authorization" request header is "Bearer jane_current_token"
    When I send a POST request to "/auth/token<query>"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {<token_in_data> "expires_in": 31622400}
      }
      """
    And the response header "Set-Cookie" should be "<expected_cookie>"
    And the table "sessions" should be:
      | session_id | user_id | refresh_token             |
      | 1          | 12      |                           |
      | 2          | 13      | jane_new_refreshtoken     |
      | 3          | 14      | john_current_refreshtoken |
    And the table "access_tokens" should be:
      | session_id | issued_at           | expires_at          | token              |
      | 1          | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | tmp_old_token      |
      | 1          | 2020-01-01 01:00:12 | 2020-01-01 03:00:12 | tmp_current_token  |
      | 2          | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | jane_old_token     |
      | 2          | 2020-01-01 01:50:00 | 2020-01-01 03:50:00 | jane_current_token |
      | 2          | 2020-01-01 02:00:00 | 2021-01-01 02:00:00 | jane_new_token     | # the new token
      | 3          | 2020-01-01 01:50:00 | 2020-01-01 03:50:00 | john_current_token |
    Examples:
      | query                            | token_in_data                     | expected_cookie                                                                                                                                               |
      |                                  | "access_token": "jane_new_token", | [Header not defined]                                                                                                                                          |
      | ?use_cookie=1&cookie_secure=1    |                                   | access_token=2!jane_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Fri, 01 Jan 2021 02:00:00 GMT; Max-Age=31622400; HttpOnly; Secure; SameSite=None |
      | ?use_cookie=1&cookie_same_site=1 |                                   | access_token=1!jane_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Fri, 01 Jan 2021 02:00:00 GMT; Max-Age=31622400; HttpOnly; SameSite=Strict       |

  Scenario Outline: >
      Accepts access_token cookie and removes it if cookie attributes differ for a normal user,
      since old tokens are used, the most recent one is returned
    Given the database table 'access_tokens' has also the following rows:
      | session_id | issued_at           | expires_at          | token           |
      | 2          | 2020-01-01 01:00:00 | 2020-01-01 03:00:00 | jane_old1_token |
      | 2          | 2020-01-01 01:00:00 | 2020-01-01 03:00:00 | jane_old2_token |
      | 2          | 2020-01-01 01:00:00 | 2020-01-01 03:00:00 | jane_old3_token |
    And the "Cookie" request header is "access_token=<token_cookie>"
    When I send a POST request to "/auth/token?use_cookie=1&cookie_secure=1"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"expires_in": 6600}
      }
      """
    And the response headers "Set-Cookie" should be:
      """
        <cookie_removal>
        access_token=2!jane_current_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 03:50:00 GMT; Max-Age=6600; HttpOnly; Secure; SameSite=None
      """
  Examples:
    | token_cookie                    | cookie_removal                                                                                                                 |
    | 1!jane_old_token!!              | access_token=; Expires=Wed, 01 Jan 2020 01:43:20 GMT; Max-Age=0; HttpOnly; SameSite=Strict                                     |
    | 2!jane_old1_token!127.0.0.1!/   |                                                                                                                                |
    | 2!jane_old2_token!!             | access_token=; Expires=Wed, 01 Jan 2020 01:43:20 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None                               |
    | 3!jane_old3_token!a.127.0.0.1!/ | access_token=; Path=/; Domain=a.127.0.0.1; Expires=Wed, 01 Jan 2020 01:43:20 GMT; Max-Age=0; HttpOnly; Secure; SameSite=Strict |

  Scenario Outline: Accepts access_token cookie and removes it if cookie attributes differ for a temporary user
    Given the generated auth key is "tmp_new_token"
    And the database table 'access_tokens' has also the following rows:
      | session_id | issued_at           | expires_at          | token          |
      | 1          | 2020-01-01 01:00:00 | 2020-01-01 03:00:00 | tmp_old1_token |
      | 1          | 2020-01-01 01:00:00 | 2020-01-01 03:00:00 | tmp_old2_token |
      | 1          | 2020-01-01 01:00:00 | 2020-01-01 03:00:00 | tmp_old3_token |
    And the "Cookie" request header is "access_token=<token_cookie>"
    When I send a POST request to "/auth/token?use_cookie=1&cookie_secure=1"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"expires_in": 3612}
      }
      """
    And the response headers "Set-Cookie" should be:
      """
        <cookie_removal>
        access_token=2!tmp_current_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 03:00:12 GMT; Max-Age=3612; HttpOnly; Secure; SameSite=None
      """
    Examples:
      | token_cookie                      | cookie_removal                                                                                                                   |
      | 2!tmp_old_token!a.127.0.0.1!/api/ | access_token=; Path=/api/; Domain=a.127.0.0.1; Expires=Wed, 01 Jan 2020 01:43:20 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None |
      | 2!tmp_old1_token!127.0.0.1!/      |                                                                                                                                  |
      | 2!tmp_old2_token!!                | access_token=; Expires=Wed, 01 Jan 2020 01:43:20 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None                                 |
      | 3!tmp_old3_token!a.127.0.0.1!/    | access_token=; Path=/; Domain=a.127.0.0.1; Expires=Wed, 01 Jan 2020 01:43:20 GMT; Max-Age=0; HttpOnly; Secure; SameSite=Strict   |

  Scenario Outline: Accepts cookie parameters from post data
    Given the generated auth key is "tmp_new_token"
    And the "Authorization" request header is "Bearer tmp_current_token"
    And the "Content-Type" request header is "<content-type>"
    When I send a POST request to "/auth/token" with the following body:
      """
      <data>
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"expires_in": 7200}
      }
      """
    And the response header "Set-Cookie" should be "<expected_cookie>"
    And logs should contain:
      """
      Refreshed a session token expiring in 7200 seconds for a temporary user with group_id = 12
      """
    And the table "sessions" should be:
      | session_id | user_id | refresh_token             |
      | 1          | 12      |                           |
      | 2          | 13      | jane_current_refreshtoken |
      | 3          | 14      | john_current_refreshtoken |
    And the table "access_tokens" should be:
      | session_id | issued_at           | expires_at          | token              |
      | 1          | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | tmp_old_token      |
      | 1          | 2020-01-01 01:00:12 | 2020-01-01 03:00:12 | tmp_current_token  |
      | 1          | 2020-01-01 02:00:00 | 2020-01-01 04:00:00 | tmp_new_token      |
      | 2          | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | jane_old_token     |
      | 2          | 2020-01-01 01:50:00 | 2020-01-01 03:50:00 | jane_current_token |
      | 3          | 2020-01-01 01:50:00 | 2020-01-01 03:50:00 | john_current_token |
    Examples:
      | content-type                      | data                                                              | expected_cookie                                                                                                                                            |
      | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=1                                      | access_token=2!tmp_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 04:00:00 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None   |
      | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=1&cookie_same_site=1                   | access_token=3!tmp_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 04:00:00 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=Strict |
      | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=0&cookie_same_site=1                   | access_token=1!tmp_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 04:00:00 GMT; Max-Age=7200; HttpOnly; SameSite=Strict         |
      | application/jsoN; charset=utf8    | {"use_cookie":true,"cookie_secure":true}                          | access_token=2!tmp_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 04:00:00 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None   |
      | application/json                  | {"use_cookie":true,"cookie_secure":true,"cookie_same_site":true}  | access_token=3!tmp_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 04:00:00 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=Strict |
      | Application/json                  | {"use_cookie":true,"cookie_secure":false,"cookie_same_site":true} | access_token=1!tmp_new_token!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 01 Jan 2020 04:00:00 GMT; Max-Age=7200; HttpOnly; SameSite=Strict         |
