Feature: Refresh an access token - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name        | type |
      | 13 | jane        | User |
    And the database has the following table 'users':
      | group_id | login       | temp_user |
      | 13       | jane        | false     |
    And the database has the following table 'sessions':
      | user_id | expires_at          | access_token              |
      | 13      | 2019-07-16 22:02:31 | anotheraccesstokenforjane |
      | 13      | 3019-07-16 22:02:29 | accesstokenforjane        |

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
    And the table "refresh_tokens" should stay unchanged

  Scenario Outline: Invalid cookie attributes
    Given the "Authorization" request header is "Bearer accesstokenforjane"
    And I send a POST request to "/auth/token<query>"
    Then the response code should be 400
    And the response error message should contain "<expected_error>"
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged
  Examples:
    | query                                            | expected_error                                                                 |
    | ?use_cookie=1                                    | One of cookie_secure and cookie_same_site must be true when use_cookie is true |
    | ?use_cookie=1&cookie_same_site=0&cookie_secure=0 | One of cookie_secure and cookie_same_site must be true when use_cookie is true |
    | ?use_cookie=abc                                  | Wrong value for use_cookie (should have a boolean value (0 or 1))              |
    | ?cookie_same_site=abc                            | Wrong value for cookie_same_site (should have a boolean value (0 or 1))        |
    | ?cookie_secure=abc                               | Wrong value for cookie_secure (should have a boolean value (0 or 1))           |
