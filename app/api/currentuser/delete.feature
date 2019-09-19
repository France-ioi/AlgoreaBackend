Feature: Delete the current user
  Background:
    Given the database has the following table 'groups':
      | id | type      | name       |
      | 1  | Base      | Root       |
      | 2  | Base      | RootSelf   |
      | 3  | Base      | RootAdmin  |
      | 4  | Base      | RootTemp   |
      | 21 | UserSelf  | user       |
      | 22 | UserAdmin | user-admin |
      | 31 | UserSelf  | tmp-1234   |
    And the database has the following table 'groups_groups':
      | group_parent_id | group_child_id |
      | 1               | 2              |
      | 1               | 3              |
      | 2               | 4              |
      | 2               | 21             |
      | 3               | 22             |
      | 4               | 31             |
    And the database has the following table 'groups_ancestors':
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 1              | true    |
      | 1                 | 2              | false   |
      | 1                 | 3              | false   |
      | 1                 | 4              | false   |
      | 1                 | 21             | false   |
      | 1                 | 22             | false   |
      | 1                 | 31             | false   |
      | 2                 | 2              | true    |
      | 2                 | 4              | false   |
      | 2                 | 21             | false   |
      | 2                 | 31             | false   |
      | 3                 | 3              | true    |
      | 3                 | 22             | false   |
      | 4                 | 4              | true    |
      | 4                 | 31             | true    |
      | 21                | 21             | true    |
      | 22                | 22             | true    |
      | 31                | 31             | true    |
    And the database has the following table 'users':
      | id | temp_user | login    | group_self_id | group_owned_id | login_id |
      | 11 | 0         | user     | 21            | 22             | 1234567  |
      | 12 | 1         | tmp-1234 | 31            | null           | null     |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
        callbackURL: "https://backend.algorea.org/auth/login-callback"
      """

  Scenario: Regular user
    Given I am the user with id "11"
    And the login module "unlink_client" endpoint for user id "1234567" returns 200 with encoded body:
      """
      {"success":true}
      """
    When I send a DELETE request to "/current-user"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "deleted"
      }
      """
    And the table "users" should be:
      | id | temp_user | login    | group_self_id | group_owned_id |
      | 12 | 1         | tmp-1234 | 31            | null           |
    And the table "groups" should be:
      | id | type     | name      |
      | 1  | Base     | Root      |
      | 2  | Base     | RootSelf  |
      | 3  | Base     | RootAdmin |
      | 4  | Base     | RootTemp  |
      | 31 | UserSelf | tmp-1234  |
    And the table "groups_groups" should be:
      | group_parent_id | group_child_id |
      | 1               | 2              |
      | 1               | 3              |
      | 2               | 4              |
      | 4               | 31             |
    And the table "groups_ancestors" should be:
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 1              | true    |
      | 1                 | 2              | false   |
      | 1                 | 3              | false   |
      | 1                 | 4              | false   |
      | 1                 | 31             | false   |
      | 2                 | 2              | true    |
      | 2                 | 4              | false   |
      | 2                 | 31             | false   |
      | 3                 | 3              | true    |
      | 4                 | 4              | true    |
      | 4                 | 31             | true    |
      | 31                | 31             | true    |

  Scenario: Temporary user
    Given I am the user with id "12"
    When I send a DELETE request to "/current-user"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "deleted"
      }
      """
    And the table "users" should be:
      | id | temp_user | login   | group_self_id | group_owned_id |
      | 11 | 0         | user    | 21            | 22             |
    And the table "groups" should be:
      | id | type      | name       |
      | 1  | Base      | Root       |
      | 2  | Base      | RootSelf   |
      | 3  | Base      | RootAdmin  |
      | 4  | Base      | RootTemp   |
      | 21 | UserSelf  | user       |
      | 22 | UserAdmin | user-admin |
    And the table "groups_groups" should be:
      | group_parent_id | group_child_id |
      | 1               | 2              |
      | 1               | 3              |
      | 2               | 4              |
      | 2               | 21             |
      | 3               | 22             |
    And the table "groups_ancestors" should be:
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 1              | true    |
      | 1                 | 2              | false   |
      | 1                 | 3              | false   |
      | 1                 | 4              | false   |
      | 1                 | 21             | false   |
      | 1                 | 22             | false   |
      | 2                 | 2              | true    |
      | 2                 | 4              | false   |
      | 2                 | 21             | false   |
      | 3                 | 3              | true    |
      | 3                 | 22             | false   |
      | 4                 | 4              | true    |
      | 21                | 21             | true    |
      | 22                | 22             | true    |
