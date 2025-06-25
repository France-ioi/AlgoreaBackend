Feature: Support for parallel sessions
  Background:
    Given there are the following groups:
      | group         | parent    | members               |
      | @NonTempUsers | @AllUsers | @User1,@UserUntouched |
    And the time now is "2020-01-01T01:00:00Z"
    And there are the following sessions:
      | session                | user           | refresh_token       |
      | @Session_User1_1       | @User1         | rt_user_1_session_1 |
      | @Session_User1_2       | @User1         | rt_user_1_session_2 |
      | @Session_User1_3       | @User1         | rt_user_1_session_3 |
      | @Session_UserUntouched | @UserUntouched | rt_user_untouched   |
    And there are the following access tokens:
      | session                | token                                    | issued_at           | expires_at          |
      | @Session_User1_1       | t_user_1_session_1_expired               | 2019-01-01 00:00:00 | 2019-01-01 02:00:00 |
      | @Session_User1_1       | t_user_1_session_1_old                   | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
      | @Session_User1_1       | t_user_1_session_1_most_recent           | 2020-01-01 00:30:00 | 2020-01-01 02:30:00 |
      | @Session_User1_2       | t_user_1_session_2_expired               | 2019-01-01 00:00:00 | 2019-01-01 02:00:00 |
      | @Session_User1_2       | t_user_1_session_2_old                   | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
      | @Session_User1_2       | t_user_1_session_2_most_recent_less_5min | 2020-01-01 00:56:00 | 2020-01-01 02:56:00 |
      | @Session_User1_3       | t_user_1_session_3_old                   | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
      | @Session_User1_3       | t_user_1_session_3_most_recent_more_5min | 2020-01-01 00:54:59 | 2020-01-01 02:54:59 |
      | @Session_UserUntouched | t_user_untouched_expired                 | 2019-01-01 00:00:00 | 2019-01-01 02:00:00 |
      | @Session_UserUntouched | t_user_untouched                         | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      """

  Scenario: Should return the most recent access token if the one used to refresh isn't the most recent one, along with the right expiration time
    When the "Authorization" request header is "Bearer t_user_1_session_1_old"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    # expires_in is 5400 seconds = 1h30
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "access_token": "t_user_1_session_1_most_recent",
          "expires_in": 5400
        }
      }
      """

  Scenario: Should return the same access token if we try to refresh a token not older than 5 minutes
    When the "Authorization" request header is "Bearer t_user_1_session_2_most_recent_less_5min"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    # expires_in is 6960 seconds = 1h56
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "access_token": "t_user_1_session_2_most_recent_less_5min",
          "expires_in": 6960
        }
      }
      """

  Scenario: Should return a new access token if we try to refresh a token older than 5 minutes
    Given the login module "token" endpoint for refresh token "rt_user_1_session_3" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":7200,
        "access_token":"t_user_1_session_3_new",
        "refresh_token":"rt_user_1_session_3_new"
      }
      """
    When the "Authorization" request header is "Bearer t_user_1_session_3_most_recent_more_5min"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "access_token": "t_user_1_session_3_new",
          "expires_in": 7200
        }
      }
      """
