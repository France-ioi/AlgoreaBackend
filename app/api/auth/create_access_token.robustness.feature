Feature: Login callback - robustness
  Scenario: Both code and Authorization header are present
    Given the "Authorization" request header is "Bearer 1234567890"
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 400
    And the response error message should contain "Only one of the 'code' query parameter and the 'Authorization' header can be given"
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
    And the table "refresh_tokens" should stay unchanged

  Scenario: The code is missing
    When I send a POST request to "/auth/token"
    Then the response code should be 400
    And the response error message should contain "Missing code"
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
