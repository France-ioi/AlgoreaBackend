Feature: Request a new access token
  Background:
    Given the database has the following table 'groups':
      | id | name        | type      |
      | 12 | tmp-1234567 | UserSelf  |
      | 13 | jane        | UserSelf  |
      | 14 | john        | UserSelf  |
    And the database has the following table 'users':
      | group_id | login       | temp_user |
      | 12       | tmp-1234567 | true      |
      | 13       | jane        | false     |
      | 14       | john        | false     |
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'sessions':
      | user_id | expires_at          | access_token              |
      | 12      | 2019-07-16 22:02:29 | someaccesstoken           |
      | 12      | 2019-07-16 22:02:40 | anotheraccesstoken        |
      | 13      | 2019-07-16 22:02:29 | accesstokenforjane        |
      | 13      | 2019-07-16 22:02:31 | anotheraccesstokenforjane |
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
        callbackURL: "https://backend.algorea.org/auth/login-callback"
      """

  Scenario: Request a new access token for a temporary user
    Given the generated auth key is "newaccesstoken"
    And the "Authorization" request header is "Bearer someaccesstoken"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"access_token": "newaccesstoken", "expires_in": 7200}
      }
      """
    And logs should contain:
      """
      Generated a session token expiring in 7200 seconds for a temporary user with group_id = 12
      """
    And the table "sessions" should be:
      | user_id | expires_at          | access_token              |
      | 12      | 2019-07-16 22:02:29 | someaccesstoken           |
      | 12      | 2019-07-17 00:02:28 | newaccesstoken            |
      | 13      | 2019-07-16 22:02:29 | accesstokenforjane        |
      | 13      | 2019-07-16 22:02:31 | anotheraccesstokenforjane |
    And the table "refresh_tokens" should stay unchanged

  Scenario: Request a new access token for a normal user
    Given the login module "token" endpoint for refresh token "refreshtokenforjane" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622400,
        "access_token":"newaccesstokenforjane",
        "refresh_token":"newrefreshtokenforjane"
      }
      """
    And the time now is "2019-07-16T22:02:29Z"
    And the "Authorization" request header is "Bearer accesstokenforjane"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"access_token": "newaccesstokenforjane", "expires_in": 31622400}
      }
      """
    And the table "sessions" should be:
      | user_id | expires_at          | access_token          |
      | 12      | 2019-07-16 22:02:29 | someaccesstoken       |
      | 12      | 2019-07-16 22:02:40 | anotheraccesstoken    |
      | 13      | 2019-07-16 22:02:29 | accesstokenforjane    |
      | 13      | 2020-07-16 22:02:29 | newaccesstokenforjane |
    And the table "refresh_tokens" should be:
      | user_id | refresh_token          |
      | 13      | newrefreshtokenforjane |
      | 14      | refreshtokenforjohn    |
