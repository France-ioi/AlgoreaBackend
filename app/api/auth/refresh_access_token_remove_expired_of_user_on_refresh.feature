Feature: Support for parallel sessions
  Background:
    Given there are the following groups:
      | group     | parent | members                                        |
      | @AllUsers |        | @User1,@UserUntouched,@UserWithoutExpiredToken |
    And the time now is "2020-01-01T01:00:00Z"
    And the DB time now is "2020-01-01 01:00:00"
    And there are the following sessions:
      | session                          | user                     | refresh_token                 |
      | @Session_User1_1                 | @User1                   | rt_user_1_session_1           |
      | @Session_User1_2                 | @User1                   | rt_user_1_session_2           |
      | @Session_User1_3                 | @User1                   | rt_user_1_session_3           |
      | @Session_UserUntouched           | @UserUntouched           | rt_user_untouched             |
      | @Session_UserWithoutExpiredToken | @UserWithoutExpiredToken | rt_user_without_expired_token |
    And there are the following access tokens:
      | session                          | token                                    | issued_at           | expires_at          |
      | @Session_User1_1                 | t_user_1_session_1_expired               | 2019-01-01 00:00:00 | 2019-01-01 02:00:00 |
      | @Session_User1_1                 | t_user_1_session_1_old                   | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
      | @Session_User1_1                 | t_user_1_session_1_most_recent           | 2020-01-01 00:30:00 | 2020-01-01 02:30:00 |
      | @Session_User1_2                 | t_user_1_session_2_expired               | 2019-01-01 00:00:00 | 2019-01-01 02:00:00 |
      | @Session_User1_2                 | t_user_1_session_2_old                   | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
      | @Session_User1_2                 | t_user_1_session_2_most_recent_less_5min | 2020-01-01 00:56:00 | 2020-01-01 02:56:00 |
      | @Session_User1_3                 | t_user_1_session_3_old                   | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
      | @Session_User1_3                 | t_user_1_session_3_most_recent_more_5min | 2020-01-01 00:54:59 | 2020-01-01 02:54:59 |
      | @Session_UserUntouched           | t_user_untouched_expired                 | 2019-01-01 00:00:00 | 2019-01-01 02:00:00 |
      | @Session_UserUntouched           | t_user_untouched                         | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
      | @Session_UserWithoutExpiredToken | t_user_without_expired_token             | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      """

  Scenario: Should remove the expired access tokens of the user when refreshing a token
    Given the login module "token" endpoint for refresh token "rt_user_1_session_1" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":7200,
        "access_token":"t_user_1_session_1_most_recent",
        "refresh_token":"rt_user_1_session_1_new"
      }
      """
    And the "Authorization" request header is "Bearer t_user_1_session_1_most_recent"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    And there are 7 access tokens for user @User1
    And there is no access token "t_user_1_session_1_expired"
    And there is no access token "t_user_1_session_2_expired"
    # UserUntouched's access tokens shouldn't be touched.
    And there are 2 access tokens for user @UserUntouched

  Scenario: Should remove the expired access tokens of the user when refreshing a token, when the user is a temp user
    Given there are the following users:
      | user                     | groups    | temp_user |
      | @User1                   | @AllUsers | true      |
      | @UserUntouched           | @AllUsers | true      |
      | @UserWithoutExpiredToken | @AllUsers | true      |
    And the "Authorization" request header is "Bearer t_user_1_session_1_most_recent"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    And there are 7 access tokens for user @User1
    And there is no access token "t_user_1_session_1_expired"
    And there is no access token "t_user_1_session_2_expired"
    # UserUntouched's access tokens shouldn't be touched.
    And there are 2 access tokens for user @UserUntouched

  Scenario: Should not remove any access token if there's no expired one for the user
    Given the login module "token" endpoint for refresh token "rt_user_without_expired_token" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":7200,
        "access_token":"t_user_without_expired_token",
        "refresh_token":"t_user_without_expired_token_new"
      }
      """
    And the "Authorization" request header is "Bearer t_user_without_expired_token"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    And there are 2 access tokens for user @UserWithoutExpiredToken

  Scenario: Should not remove any access token if there's no expired one for the user, when the user is a temp user
    Given there are the following users:
      | user                     | groups    | temp_user |
      | @User1                   | @AllUsers | true      |
      | @UserUntouched           | @AllUsers | true      |
      | @UserWithoutExpiredToken | @AllUsers | true      |
    And the "Authorization" request header is "Bearer t_user_without_expired_token"
    When I send a POST request to "/auth/token"
    Then the response code should be 201
    And there are 2 access tokens for user @UserWithoutExpiredToken
