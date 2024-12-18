Feature: Refresh an access token - robustness
  Background:
    Given the DB time now is "2020-01-01 01:00:00"
    And the database has the following user:
      | group_id | login |
      | 13       | jane  |
    And the database has the following table "sessions":
      | session_id | user_id | refresh_token |
      | 1          | 13      |               |
    And the database has the following table "access_tokens":
      | session_id | issued_at           | expires_at          | token              |
      | 1          | 2019-01-01 00:00:00 | 2019-01-01 02:00:00 | jane_expired_token |
      | 1          | 2020-01-01 00:00:00 | 2020-01-01 02:00:00 | jane_current_token |

  Scenario: No refresh token in the DB
    Given the "Authorization" request header is "Bearer jane_current_token"
    When I send a POST request to "/auth/token"
    Then the response code should be 404
    And the response error message should contain "No refresh token found in the DB for the authenticated user"
    And logs should contain:
      """
      No refresh token found in the DB for user 13
      """
    And the table "sessions" should stay unchanged
    # The expired token has been removed
    And the table "access_tokens" should stay unchanged

  Scenario: Should return an error when trying to refresh an expired token
    Given the "Authorization" request header is "Bearer jane_expired_token"
    When I send a POST request to "/auth/token"
    Then the response code should be 401
