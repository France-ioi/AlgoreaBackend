Feature: Login callback - robustness
  Scenario: Both code and Authorization header are present
    Given the "Authorization" request header is "Bearer 1234567890"
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 400
    And the response error message should contain "Only one of the 'code' parameter and the 'Authorization' header can be given"
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario: The code is missing for an unauthenticated user
    When I send a POST request to "/auth/token"
    Then the response code should be 401
    And the response error message should contain "No access token provided"
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario: Invalid JSON data
    Given the "Content-Type" request header is "application/json"
    When I send a POST request to "/auth/token" with the following body:
    """
    code=1234
    """
    Then the response code should be 400
    And the response error message should contain "Invalid character 'c' looking for beginning of value"
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario: Invalid form data
    Given the "Content-Type" request header is "application/x-www-form-urlencoded"
    When I send a POST request to "/auth/token" with the following body:
    """
    %%%%
    """
    Then the response code should be 400
    And the response error message should contain "Invalid URL escape "%%%""
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario: Invalid request content type
    Given the "Content-Type" request header is "application/xml"
    When I send a POST request to "/auth/token" with the following body:
    """
    <code>1234</code>
    """
    Then the response code should be 415
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario: OAuth error
    Given the DB time now is "2019-07-16 22:02:28"
    And the login module "token" endpoint for code "somecode" returns 500 with body:
      """
      Unknown error
      """
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 500
    And the response error message should contain "Oauth2: cannot fetch token: 500"
    And the response error message should contain "Response: Unknown error"
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario: User API error
    Given the DB time now is "2019-07-16 22:02:28"
    And the login module "token" endpoint for code "somecode" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622420,
        "access_token":"accesstoken",
        "refresh_token":"refreshtoken"
      }
      """
    And the login module "account" endpoint for token "accesstoken" returns 500 with body:
      """
      Unknown error
      """
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 500
    And the response error message should contain "Can't retrieve user's profile (status code = 500)"
    And logs should contain:
      """
      Can't retrieve user's profile (status code = 500, response = "Unknown error")
      """
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario: User profile can't be parsed
    Given the DB time now is "2019-07-16 22:02:28"
    And the login module "token" endpoint for code "somecode" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622420,
        "access_token":"accesstoken",
        "refresh_token":"refreshtoken"
      }
      """
    And the login module "account" endpoint for token "accesstoken" returns 200 with body:
      """
      Not a JSON
      """
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 500
    And the response error message should contain "Can't parse user's profile"
    And logs should contain:
      """
      Can't parse user's profile (response = "Not a JSON", error = "invalid character 'N' looking for beginning of value")
      """
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario Outline: User profile is invalid
    Given the DB time now is "2019-07-16 22:02:28"
    And the login module "token" endpoint for code "somecode" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622420,
        "access_token":"accesstoken",
        "refresh_token":"refreshtoken"
      }
      """
    And the login module "account" endpoint for token "accesstoken" returns 200 with body:
      """
      <profile_body>
      """
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 500
    And the response error message should contain "User's profile is invalid"
    And logs should contain:
      """
      User's profile is invalid (response = "{{`<profile_body>`|safeJs}}", error = "<error_text>")
      """
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged
  Examples:
    | profile_body      | error_text                 |
    | {"login":"login"} | no id in user's profile    |
    | {"id":12345}      | no login in user's profile |

  Scenario Outline: Invalid cookie attributes
    Given I send a POST request to "/auth/token<query>"
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
