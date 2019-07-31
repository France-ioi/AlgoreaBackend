Feature: Request a new access token
  Background:
    Given the database has the following table 'users':
      | ID | sLogin      | tempUser |
      | 2  | tmp-1234567 | true     |
      | 3  | jane        | false    |
      | 4  | john        | false    |
    And the DB time now is "2019-07-16T22:02:28Z"
    And the database has the following table 'sessions':
      | idUser | sExpirationDate      | sAccessToken              |
      | 2      | 2019-07-16T22:02:29Z | someaccesstoken           |
      | 2      | 2019-07-16T22:02:40Z | anotheraccesstoken        |
      | 3      | 2019-07-16T22:02:29Z | accesstokenforjane        |
      | 3      | 2019-07-16T22:02:31Z | anotheraccesstokenforjane |
    And the database has the following table 'refresh_tokens':
      | idUser | sRefreshToken       |
      | 3      | refreshtokenforjane |
      | 4      | refreshtokenforjohn |
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
      Generated a session token expiring in 7200 seconds for a temporary user 2
      """
    And the table "sessions" should be:
      | idUser | sExpirationDate      | sAccessToken              |
      | 2      | 2019-07-16T22:02:29Z | someaccesstoken           |
      | 2      | 2019-07-17T00:02:28Z | newaccesstoken            |
      | 3      | 2019-07-16T22:02:29Z | accesstokenforjane        |
      | 3      | 2019-07-16T22:02:31Z | anotheraccesstokenforjane |
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
      | idUser | sExpirationDate      | sAccessToken          |
      | 2      | 2019-07-16T22:02:29Z | someaccesstoken       |
      | 2      | 2019-07-16T22:02:40Z | anotheraccesstoken    |
      | 3      | 2019-07-16T22:02:29Z | accesstokenforjane    |
      | 3      | 2020-07-16T22:02:29Z | newaccesstokenforjane |
    And the table "refresh_tokens" should be:
      | idUser | sRefreshToken          |
      | 3      | newrefreshtokenforjane |
      | 4      | refreshtokenforjohn    |
