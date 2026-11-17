Feature: Create a user and sign in
  Background:
    Given the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: 2
          nonTempUsersGroup: 3
          tempUsersGroup: 4
      """
    And the database has the following table "groups":
      | id | name         | type | created_at          |
      | 2  | AllUsers     | Base | 2020-01-01 00:00:00 |
      | 3  | NonTempUsers | Base | 2020-01-01 00:00:00 |
      | 4  | TempUsers    | Base | 2020-01-01 00:00:00 |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 2               | 3              |
      | 2               | 4              |

  Scenario: A visitor with no session gets a user and a session cookie
    Given the time now is "2019-07-16T22:02:28Z"
    And the login module "create_user" endpoint returns 200 with encoded body:
      """
      {"success": true, "data": {"code": "bridgecode", "redirect_uri": "https://contest.example/cb", "password": "should-not-leak"}}
      """
    And the login module "token" endpoint for code "bridgecode" and redirect_uri "https://contest.example/cb" returns 200 with body:
      """
      {
        "token_type": "Bearer",
        "expires_in": 7200,
        "access_token": "useraccesstoken",
        "refresh_token": "userrefreshtoken"
      }
      """
    And the login module "account" endpoint for token "useraccesstoken" returns 200 with body:
      """
      {"id": 100000042, "login": "beoi27_k3mnpq", "language": "en", "badges": []}
      """
    # The participation code must not be forwarded: the façade stub expects an empty body.
    When I send a POST request to "/auth/user?use_cookie=1&cookie_secure=1&code=join-me"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "login": "beoi27_k3mnpq",
          "expires_in": 7200
        }
      }
      """
    And the response header "Set-Cookie" should be "access_token=3!useraccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=Strict"
    And an event "user_authenticated" should have been dispatched with:
      """
      {"user_id": "5577006791947779410", "login_id": "100000042", "login": "beoi27_k3mnpq", "user_ip": "127.0.0.1", "profile": null}
      """
    And the table "users" should be:
      | group_id            | latest_login_at     | latest_activity_at  | temp_user | registered_at       | latest_profile_sync_at | login_id  | login          | default_language | last_ip   |
      | 5577006791947779410 | 2019-07-16 22:02:28 | 2019-07-16 22:02:28 | 0         | 2019-07-16 22:02:28 | 2019-07-16 22:02:28    | 100000042 | beoi27_k3mnpq  | en               | 127.0.0.1 |
    And the table "groups" should be:
      | id                  | name           | type | description    | created_at          | is_open | send_emails |
      | 2                   | AllUsers       | Base | null           | 2020-01-01 00:00:00 | false   | false       |
      | 3                   | NonTempUsers   | Base | null           | 2020-01-01 00:00:00 | false   | false       |
      | 4                   | TempUsers      | Base | null           | 2020-01-01 00:00:00 | false   | false       |
      | 5577006791947779410 | beoi27_k3mnpq  | User | beoi27_k3mnpq  | 2019-07-16 22:02:28 | false   | false       |
    And the table "groups_groups" should be:
      | parent_group_id     | child_group_id      |
      | 2                   | 3                   |
      | 2                   | 4                   |
      | 3                   | 5577006791947779410 |
    And the table "sessions" should be:
      | session_id          | user_id             | refresh_token           |
      | 8674665223082153551 | 5577006791947779410 | userrefreshtoken  |
    And the table "access_tokens" should be:
      | session_id          | token                  | expires_at          | issued_at           |
      | 8674665223082153551 | useraccesstoken  | 2019-07-17 00:02:28 | 2019-07-16 22:02:28 |

  Scenario: A temporary session is replaced by the user session
    Given the time now is "2019-07-16T22:02:28Z"
    And the database has the following users:
      | group_id | login  | temp_user |
      | 22       | tmp-22 | true      |
    And the database has the following table "sessions":
      | session_id | user_id | refresh_token |
      | 900        | 22      | temprefresh   |
    And the database has the following table "access_tokens":
      | session_id | token     | issued_at           | expires_at          |
      | 900        | temptoken | 2019-07-16 20:00:00 | 2019-07-17 00:00:00 |
    And the "Cookie" request header is "access_token=3!temptoken!127.0.0.1!/"
    And the login module "create_user" endpoint returns 200 with encoded body:
      """
      {"success": true, "data": {"code": "bridgecode", "redirect_uri": "https://contest.example/cb", "password": "should-not-leak"}}
      """
    And the login module "token" endpoint for code "bridgecode" and redirect_uri "https://contest.example/cb" returns 200 with body:
      """
      {
        "token_type": "Bearer",
        "expires_in": 7200,
        "access_token": "useraccesstoken",
        "refresh_token": "userrefreshtoken"
      }
      """
    And the login module "account" endpoint for token "useraccesstoken" returns 200 with body:
      """
      {"id": 100000042, "login": "beoi27_k3mnpq", "language": "en", "badges": []}
      """
    When I send a POST request to "/auth/user?use_cookie=1&cookie_secure=1"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "login": "beoi27_k3mnpq",
          "expires_in": 7200
        }
      }
      """
    And the response header "Set-Cookie" should be "access_token=3!useraccesstoken!127.0.0.1!/; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:28 GMT; Max-Age=7200; HttpOnly; Secure; SameSite=Strict"
    And the table "users" at group_id "22" should be:
      | group_id | login  | temp_user |
      | 22       | tmp-22 | 1         |
    And the table "users" at group_id "5577006791947779410" should be:
      | group_id            | login         | temp_user |
      | 5577006791947779410 | beoi27_k3mnpq | 0         |
    And the table "sessions" at user_id "5577006791947779410" should be:
      | session_id          | user_id             | refresh_token          |
      | 8674665223082153551 | 5577006791947779410 | userrefreshtoken |
    And the table "sessions" at user_id "22" should be:
      | session_id | user_id | refresh_token |
      | 900        | 22      | temprefresh   |
