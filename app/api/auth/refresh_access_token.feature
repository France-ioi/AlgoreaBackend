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
    And the time now is "2019-07-16T22:02:28Z"
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'sessions':
      | session_id | user_id | refresh_token       |
      | 1          | 12      |                     |
      | 2          | 13      | refreshtokenforjane |
      | 3          | 14      | refreshtokenforjohn |
    And the database has the following table 'access_tokens':
      | session_id | token                     | expires_at          |
      | 1          | someaccesstoken           | 2019-07-16 22:02:29 |
      | 1          | anotheraccesstoken        | 2019-07-16 22:02:40 |
      | 2          | accesstokenforjane        | 2019-07-16 22:02:29 |
      | 2          | anotheraccesstokenforjane | 2019-07-16 22:02:31 |
      | 3          | accesstokenjohn           | 2019-07-16 22:02:31 |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      """

  Scenario Outline: Request a new access token for a temporary user
    Given the generated auth key is "newaccesstoken"
    And the "Authorization" request header is "Bearer anotheraccesstoken"
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
      Generated a session token expiring in 7200 seconds for a temporary user with group_id = 12
      """
    And the table "sessions" should be:
      | session_id          | user_id | refresh_token       |
      | 2                   | 13      | refreshtokenforjane |
      | 3                   | 14      | refreshtokenforjohn |
      | 5577006791947779410 | 12      | null                |
    And the table "access_tokens" should be:
      | session_id          | token                     | expires_at          |
      | 2                   | accesstokenforjane        | 2019-07-16 22:02:29 |
      | 2                   | anotheraccesstokenforjane | 2019-07-16 22:02:31 |
      | 3                   | accesstokenjohn           | 2019-07-16 22:02:31 |
      | 5577006791947779410 | newaccesstoken            | 2019-07-17 00:02:28 |
  Examples:
    | query                            | current_cookie        | token_in_data                    | expected_cookie                                                                                                                                           |
    |                                  | [NULL]                | "access_token":"newaccesstoken", | [NULL]                                                                                                                                                    |
    | ?use_cookie=1&cookie_secure=1    | [NULL]                |                                  | access_token=2!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None |
    | ?use_cookie=1&cookie_same_site=1 | [NULL]                |                                  | access_token=1!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; SameSite=Strict       |
    | ?use_cookie=0                    | access_token=0!1234!! | "access_token":"newaccesstoken", | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; SameSite=None                                                                  |

  Scenario Outline: Request a new access token for a normal user
    Given the login module "token" endpoint for refresh token "refreshtokenforjane" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622400,
        "access_token":"newaccesstokenforjane",
        "refresh_token":"newrefreshtokenforjane"
      }
      """
    And the "Authorization" request header is "Bearer accesstokenforjane"
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
      | session_id | user_id | refresh_token          |
      | 1          | 12      |                        |
      | 2          | 13      | newrefreshtokenforjane |
      | 3          | 14      | refreshtokenforjohn    |
    And the table "access_tokens" should be:
      | session_id | token                 | expires_at          |
      | 1          | anotheraccesstoken    | 2019-07-16 22:02:40 |
      | 1          | someaccesstoken       | 2019-07-16 22:02:29 |
      | 2          | accesstokenforjane    | 2019-07-16 22:02:29 |
      | 2          | newaccesstokenforjane | 2020-07-16 22:02:28 |
      | 3          | accesstokenjohn       | 2019-07-16 22:02:31 |
  Examples:
    | query                            | token_in_data                            | expected_cookie                                                                                                                                                      |
    |                                  | "access_token": "newaccesstokenforjane", | [NULL]                                                                                                                                                               |
    | ?use_cookie=1&cookie_secure=1    |                                          | access_token=2!newaccesstokenforjane!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Thu, 16 Jul 2020 22:02:28 GMT; Max-Age=31622400; HttpOnly; Secure; SameSite=None |
    | ?use_cookie=1&cookie_same_site=1 |                                          | access_token=1!newaccesstokenforjane!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Thu, 16 Jul 2020 22:02:28 GMT; Max-Age=31622400; HttpOnly; SameSite=Strict       |

  Scenario Outline: Accepts access_token cookie and removes it if cookie attributes differ for a normal user
    Given the database table 'access_tokens' has also the following rows:
      | session_id | expires_at          | token                     |
      | 2          | 2019-07-16 22:02:31 | onemoreaccesstokenforjane |
      | 2          | 2019-07-16 22:02:31 | andmoreaccesstokenforjane |
      | 2          | 2019-07-16 22:02:31 | moremoraccesstokenforjane |
  And the login module "token" endpoint for refresh token "refreshtokenforjane" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622400,
        "access_token":"newaccesstoken",
        "refresh_token":"newrefreshtoken"
      }
      """
    And the "Cookie" request header is "access_token=<token_cookie>"
    When I send a POST request to "/auth/token?use_cookie=1&cookie_secure=1"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"expires_in": 31622400}
      }
      """
    And the response headers "Set-Cookie" should be:
      """
        <cookie_removal>
        access_token=2!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Thu, 16 Jul 2020 22:02:28 GMT; Max-Age=31622400; HttpOnly; Secure; SameSite=None
      """
  Examples:
    | token_cookie                              | cookie_removal                                                                                                                 |
    | 1!accesstokenforjane!!                    | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; SameSite=Strict                                     |
    | 2!onemoreaccesstokenforjane!127.0.0.1!/   |                                                                                                                                |
    | 2!andmoreaccesstokenforjane!!             | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None                               |
    | 3!moremoraccesstokenforjane!a.127.0.0.1!/ | access_token=; Path=/; Domain=a.127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=Strict |

  Scenario Outline: Accepts access_token cookie and removes it if cookie attributes differ for a temporary user
    Given the generated auth key is "newaccesstoken"
    And the database table 'access_tokens' has also the following rows:
      | session_id | expires_at          | token              |
      | 1          | 2019-07-16 22:02:31 | onemoreaccesstoken |
      | 1          | 2019-07-16 22:02:31 | andmoreaccesstoken |
      | 1          | 2019-07-16 22:02:31 | moremoraccesstoken |
    And the "Cookie" request header is "access_token=<token_cookie>"
    When I send a POST request to "/auth/token?use_cookie=1&cookie_secure=1"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"expires_in": 7200}
      }
      """
    And the response headers "Set-Cookie" should be:
      """
        <cookie_removal>
        access_token=2!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None
      """
    Examples:
      | token_cookie                        | cookie_removal                                                                                                                   |
      | 2!someaccesstoken!a.127.0.0.1!/api/ | access_token=; Path=/api/; Domain=a.127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None |
      | 2!onemoreaccesstoken!127.0.0.1!/    |                                                                                                                                  |
      | 2!andmoreaccesstoken!!              | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None                                 |
      | 3!moremoraccesstoken!a.127.0.0.1!/  | access_token=; Path=/; Domain=a.127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=Strict   |

  Scenario Outline: Accepts cookie parameters from post data
    Given the generated auth key is "newaccesstoken"
    And the "Authorization" request header is "Bearer anotheraccesstoken"
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
      Generated a session token expiring in 7200 seconds for a temporary user with group_id = 12
      """
    And the table "sessions" should be:
      | session_id          | user_id | refresh_token       |
      | 2                   | 13      | refreshtokenforjane |
      | 3                   | 14      | refreshtokenforjohn |
      | 5577006791947779410 | 12      | null                |
    And the table "access_tokens" should be:
      | session_id          | token                     | expires_at          |
      | 2                   | accesstokenforjane        | 2019-07-16 22:02:29 |
      | 2                   | anotheraccesstokenforjane | 2019-07-16 22:02:31 |
      | 3                   | accesstokenjohn           | 2019-07-16 22:02:31 |
      | 5577006791947779410 | newaccesstoken            | 2019-07-17 00:02:28 |
    Examples:
    | content-type                      | data                                                              | expected_cookie                                                                                                                                             |
    | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=1                                      | access_token=2!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None   |
    | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=1&cookie_same_site=1                   | access_token=3!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=Strict |
    | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=0&cookie_same_site=1                   | access_token=1!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; SameSite=Strict         |
    | application/jsoN; charset=utf8    | {"use_cookie":true,"cookie_secure":true}                          | access_token=2!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None   |
    | application/json                  | {"use_cookie":true,"cookie_secure":true,"cookie_same_site":true}  | access_token=3!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=Strict |
    | Application/json                  | {"use_cookie":true,"cookie_secure":false,"cookie_same_site":true} | access_token=1!newaccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; SameSite=Strict         |
