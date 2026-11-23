Feature: Create a user and sign in - robustness
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
    And the database has the following table "users":
      | group_id |
    And the database has the following table "sessions":
      | session_id |
    And the database has the following table "access_tokens":
      | session_id |

  Scenario: The login module refuses to create the user
    Given the login module "create_user" endpoint returns 200 with encoded body:
      """
      {"success": false, "error": "some error"}
      """
    When I send a POST request to "/auth/user?use_cookie=1"
    Then the response code should be 500
    And the response error message should contain "Login module failed"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged
    And logs should contain:
      """
      The login module returned an error for /platform_api/accounts_manager/create_user: some error
      """

  Scenario: The login module is unavailable
    Given the login module "create_user" endpoint returns 500 with encoded body:
      """
      down
      """
    When I send a POST request to "/auth/user?use_cookie=1"
    Then the response code should be 500
    And the response error message should contain "Can't create a user"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: The login module returns a bridge that cannot be read
    Given the login module "create_user" endpoint returns 200 with encoded body:
      """
      {"success": true, "data": []}
      """
    When I send a POST request to "/auth/user?use_cookie=1"
    Then the response code should be 500
    And the response error message should contain "Can't create a user"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged

  Scenario: The login module returns no bridge code
    Given the login module "create_user" endpoint returns 200 with encoded body:
      """
      {"success": true, "data": {"redirect_uri": "https://contest.example/cb"}}
      """
    When I send a POST request to "/auth/user?use_cookie=1"
    Then the response code should be 500
    And the response error message should contain "Login module failed"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged

  Scenario: The login module returns no redirect for the bridge code
    Given the login module "create_user" endpoint returns 200 with encoded body:
      """
      {"success": true, "data": {"code": "bridgecode"}}
      """
    When I send a POST request to "/auth/user?use_cookie=1"
    Then the response code should be 500
    And the response error message should contain "Login module failed"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged

  Scenario: The bridge code cannot be exchanged
    Given the login module "create_user" endpoint returns 200 with encoded body:
      """
      {"success": true, "data": {"code": "bridgecode", "redirect_uri": "https://contest.example/cb"}}
      """
    And the login module "token" endpoint for code "bridgecode" and redirect_uri "https://contest.example/cb" returns 500 with body:
      """
      Unknown error
      """
    When I send a POST request to "/auth/user?use_cookie=1"
    Then the response code should be 500
    And the response error message should contain "Unknown error"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: The cookie flag is not a boolean
    When I send a POST request to "/auth/user?use_cookie=2"
    Then the response code should be 400
    And the response error message should contain "Wrong value for use_cookie"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: default_language is too long
    When I send a POST request to "/auth/user?default_language=russian"
    Then the response code should be 400
    And the response error message should contain "The length of default_language should be no more than 3 characters"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged
