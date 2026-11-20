Feature: Update the current user's profile - robustness
  Background:
    Given the database has the following users:
      | group_id | temp_user | login    | login_id  | profile                    |
      | 11       | 0         | mohammed | 100000001 | {"first_name": "Mohammed"} |
      | 12       | 1         | tmp-1234 | null      | null                       |

  Scenario: The route does not exist while the setting is off
    Given the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
      """
    And I am the user with id "11"
    When I send a PUT request to "/current-user/profile" with the following body:
      """
      {"profile": {"gender": "female"}}
      """
    Then the response code should be 404
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Not Found"
      }
      """
    And the table "users" should remain unchanged

  Scenario: A temporary user is refused
    Given the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        profileUpdateEnabled: true
      """
    And I am the user with id "12"
    When I send a PUT request to "/current-user/profile" with the following body:
      """
      {"profile": {"gender": "female"}}
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users" should remain unchanged

  Scenario Outline: An invalid body is refused
    Given the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        profileUpdateEnabled: true
      """
    And I am the user with id "11"
    When I send a PUT request to "/current-user/profile" with the following body:
      """
      <body>
      """
    Then the response code should be 400
    And the response error message should contain "<error>"
    And the table "users" should remain unchanged
  Examples:
    | body                           | error                 |
    | {}                             | Invalid input data    |
    | {"profile": null}              | Invalid input data    |
    | {"profile": {"gender": 1}}     | Invalid input data    |
    | {"profile": {}, "other": 1}    | Invalid input data    |
    | not json                       | Invalid input JSON    |
