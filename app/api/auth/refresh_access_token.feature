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
      | user_id | expires_at          | access_token              | use_cookie | cookie_secure | cookie_same_site | cookie_domain | cookie_path |
      | 12      | 2019-07-16 22:02:29 | someaccesstoken           | true       | true          | false            | a.127.0.0.1   | /api/       |
      | 12      | 2019-07-16 22:02:40 | anotheraccesstoken        | false      | false         | false            | null          | null        |
      | 13      | 2019-07-16 22:02:29 | accesstokenforjane        | true       | false         | true             | null          | null        |
      | 13      | 2019-07-16 22:02:31 | anotheraccesstokenforjane | false      | false         | false            | null          | null        |
    And the database has the following table 'refresh_tokens':
      | user_id | refresh_token       |
      | 13      | refreshtokenforjane |
      | 14      | refreshtokenforjohn |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      """

  Scenario Outline: Request a new access token for a temporary user
    Given the generated auth key is "newaccesstoken"
    And the "Authorization" request header is "Bearer someaccesstoken"
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
      | user_id | expires_at          | access_token              | use_cookie   | cookie_secure   | cookie_same_site   | cookie_domain   | cookie_path   |
      | 12      | 2019-07-16 22:02:29 | someaccesstoken           | true         | true            | false              | a.127.0.0.1     | /api/         |
      | 12      | 2019-07-17 00:02:28 | newaccesstoken            | <use_cookie> | <cookie_secure> | <cookie_same_site> | <cookie_domain> | <cookie_path> |
      | 13      | 2019-07-16 22:02:29 | accesstokenforjane        | true         | false           | true               | null            | null          |
      | 13      | 2019-07-16 22:02:31 | anotheraccesstokenforjane | false        | false           | false              | null            | null          |
    And the table "refresh_tokens" should stay unchanged
  Examples:
    | query                            | token_in_data                    | expected_cookie                                                                                                                             | use_cookie | cookie_secure | cookie_same_site | cookie_domain | cookie_path |
    |                                  | "access_token":"newaccesstoken", | [NULL]                                                                                                                                      | false      | false         | false            | null          | null        |
    | ?use_cookie=1&cookie_secure=1    |                                  | access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None | true       | true          | false            | 127.0.0.1     | /           |
    | ?use_cookie=1&cookie_same_site=1 |                                  | access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; SameSite=Strict       | true       | false         | true             | 127.0.0.1     | /           |

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
      | user_id | expires_at          | access_token          | use_cookie   | cookie_secure   | cookie_same_site   | cookie_domain   | cookie_path   |
      | 12      | 2019-07-16 22:02:29 | someaccesstoken       | true         | true            | false              | a.127.0.0.1     | /api/         |
      | 12      | 2019-07-16 22:02:40 | anotheraccesstoken    | false        | false           | false              | null            | null          |
      | 13      | 2019-07-16 22:02:29 | accesstokenforjane    | true         | false           | true               | null            | null          |
      | 13      | 2020-07-16 22:02:28 | newaccesstokenforjane | <use_cookie> | <cookie_secure> | <cookie_same_site> | <cookie_domain> | <cookie_path> |
    And the table "refresh_tokens" should be:
      | user_id | refresh_token          |
      | 13      | newrefreshtokenforjane |
      | 14      | refreshtokenforjohn    |
  Examples:
    | query                            | token_in_data                            | expected_cookie                                                                                                                                        | use_cookie | cookie_secure | cookie_same_site | cookie_domain | cookie_path |
    |                                  | "access_token": "newaccesstokenforjane", | [NULL]                                                                                                                                                 | false      | false         | false            | null          | null        |
    | ?use_cookie=1&cookie_secure=1    |                                          | access_token=newaccesstokenforjane; Path=/; Domain=127.0.0.1; Expires=Thu, 16 Jul 2020 22:02:28 GMT; Max-Age=31622400; HttpOnly; Secure; SameSite=None | true       | true          | false            | 127.0.0.1     | /           |
    | ?use_cookie=1&cookie_same_site=1 |                                          | access_token=newaccesstokenforjane; Path=/; Domain=127.0.0.1; Expires=Thu, 16 Jul 2020 22:02:28 GMT; Max-Age=31622400; HttpOnly; SameSite=Strict       | true       | false         | true             | 127.0.0.1     | /           |

  Scenario Outline: Accepts access_token cookie and removes it if cookie attributes differ for a normal user
    Given the database table 'sessions' has also the following rows:
      | user_id | expires_at          | access_token              | use_cookie | cookie_secure | cookie_same_site | cookie_domain | cookie_path |
      | 13      | 2019-07-16 22:02:31 | onemoreaccesstokenforjane | true       | true          | false            | 127.0.0.1     | /           |
      | 13      | 2019-07-16 22:02:31 | andmoreaccesstokenforjane | true       | true          | false            | null          | null        |
      | 13      | 2019-07-16 22:02:31 | moremoraccesstokenforjane | true       | true          | true             | a.127.0.0.1   | /           |
    And the login module "token" endpoint for refresh token "refreshtokenforjane" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622400,
        "access_token":"newaccesstoken",
        "refresh_token":"newrefreshtoken"
      }
      """
    And the "Cookie" request header is "access_token=<token>"
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
        access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Thu, 16 Jul 2020 22:02:28 GMT; Max-Age=31622400; HttpOnly; Secure; SameSite=None
      """
  Examples:
    | token                     | cookie_removal                                                                                                                 |
    | accesstokenforjane        | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; SameSite=Strict                                     |
    | anotheraccesstokenforjane |                                                                                                                                |
    | onemoreaccesstokenforjane |                                                                                                                                |
    | andmoreaccesstokenforjane | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None                               |
    | moremoraccesstokenforjane | access_token=; Path=/; Domain=a.127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=Strict |

  Scenario Outline: Accepts access_token cookie and removes it if cookie attributes differ for a temporary user
    Given the generated auth key is "newaccesstoken"
    And the database table 'sessions' has also the following rows:
      | user_id | expires_at          | access_token       | use_cookie | cookie_secure | cookie_same_site | cookie_domain | cookie_path |
      | 12      | 2019-07-16 22:02:31 | onemoreaccesstoken | true       | true          | false            | 127.0.0.1     | /           |
      | 12      | 2019-07-16 22:02:31 | andmoreaccesstoken | true       | true          | false            | null          | null        |
      | 12      | 2019-07-16 22:02:31 | moremoraccesstoken | true       | true          | true             | a.127.0.0.1   | /           |
    And the "Cookie" request header is "access_token=<token>"
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
        access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None
      """
    Examples:
      | token              | cookie_removal                                                                                                                   |
      | someaccesstoken    | access_token=; Path=/api/; Domain=a.127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None |
      | anotheraccesstoken |                                                                                                                                  |
      | onemoreaccesstoken |                                                                                                                                  |
      | andmoreaccesstoken | access_token=; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=None                                 |
      | moremoraccesstoken | access_token=; Path=/; Domain=a.127.0.0.1; Expires=Tue, 16 Jul 2019 21:45:48 GMT; Max-Age=0; HttpOnly; Secure; SameSite=Strict   |

  Scenario Outline: Accepts cookie parameters from post data
    Given the generated auth key is "newaccesstoken"
    And the "Authorization" request header is "Bearer someaccesstoken"
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
      | user_id | expires_at          | access_token              | use_cookie | cookie_secure   | cookie_same_site   | cookie_domain | cookie_path |
      | 12      | 2019-07-16 22:02:29 | someaccesstoken           | true       | true            | false              | a.127.0.0.1   | /api/       |
      | 12      | 2019-07-17 00:02:28 | newaccesstoken            | true       | <cookie_secure> | <cookie_same_site> | 127.0.0.1     | /           |
      | 13      | 2019-07-16 22:02:29 | accesstokenforjane        | true       | false           | true               | null          | null        |
      | 13      | 2019-07-16 22:02:31 | anotheraccesstokenforjane | false      | false           | false              | null          | null        |
    And the table "refresh_tokens" should stay unchanged
  Examples:
    | content-type                      | data                                                              | expected_cookie                                                                                                                               | cookie_secure | cookie_same_site |
    | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=1                                      | access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None   | true          | false            |
    | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=1&cookie_same_site=1                   | access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=Strict | true          | true             |
    | Application/x-www-form-urlencoded | use_cookie=1&cookie_secure=0&cookie_same_site=1                   | access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; SameSite=Strict         | false         | true             |
    | application/jsoN; charset=utf8    | {"use_cookie":true,"cookie_secure":true}                          | access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=None   | true          | false            |
    | application/json                  | {"use_cookie":true,"cookie_secure":true,"cookie_same_site":true}  | access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=Strict | true          | true             |
    | Application/json                  | {"use_cookie":true,"cookie_secure":false,"cookie_same_site":true} | access_token=newaccesstoken; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; SameSite=Strict         | false         | true             |
