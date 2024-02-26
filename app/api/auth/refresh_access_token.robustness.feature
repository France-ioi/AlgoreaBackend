Feature: Refresh an access token - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name        | type |
      | 13 | jane        | User |
    And the database has the following table 'users':
      | group_id | login       | temp_user |
      | 13       | jane        | false     |
    And the database has the following table 'sessions':
      | session_id | user_id | refresh_token |
      | 1          | 13      |               |
    And the database has the following table 'access_tokens':
      | session_id | token                     | expires_at          |
      | 1          | accesstokenforjane        | 3019-07-16 22:02:29 |
      | 1          | anotheraccesstokenforjane | 2019-07-16 22:02:31 |

  Scenario: No refresh token in the DB
    Given the "Authorization" request header is "Bearer accesstokenforjane"
    When I send a POST request to "/auth/token"
    Then the response code should be 404
    And the response error message should contain "No refresh token found in the DB for the authenticated user"
    And logs should contain:
      """
      No refresh token found in the DB for user 13
      """
    And the table "sessions" should stay unchanged
    And the table "access_tokens" should stay unchanged
