Feature: Delete the current user
  Background:
    Given the database has the following table 'groups':
      | id  | type     | name       | require_lock_membership_approval_until |
      | 1   | Base     | Root       | null                                   |
      | 2   | Base     | RootSelf   | 9999-12-31 23:59:59                    |
      | 4   | Base     | RootTemp   | null                                   |
      | 21  | UserSelf | user       | null                                   |
      | 31  | UserSelf | tmp-1234   | null                                   |
      | 100 | Class    | Some class | null                                   |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 2              |
      | 2               | 4              |
      | 2               | 21             |
      | 4               | 31             |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | at                  |
      | 100      | 21        | 2019-05-30 11:00:00 |
      | 100      | 31        | 2019-05-30 11:00:00 |
    And the database has the following table 'group_membership_changes':
      | group_id | member_id | at                  |
      | 100      | 21        | 2019-05-30 11:00:00 |
      | 100      | 31        | 2019-05-30 11:00:00 |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 1                 | 1              |
      | 1                 | 2              |
      | 1                 | 4              |
      | 1                 | 21             |
      | 1                 | 31             |
      | 2                 | 2              |
      | 2                 | 4              |
      | 2                 | 21             |
      | 2                 | 31             |
      | 4                 | 4              |
      | 4                 | 31             |
      | 21                | 21             |
      | 31                | 31             |
      | 100               | 100            |
    And the database has the following table 'users':
      | temp_user | login    | group_id | login_id |
      | 0         | user     | 21       | 1234567  |
      | 1         | tmp-1234 | 31       | null     |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
        callbackURL: "https://backend.algorea.org/auth/login-callback"
      """

  Scenario: Regular user
    Given I am the user with id "21"
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
      | temp_user | login    | group_id |
      | 1         | tmp-1234 | 31       |
    And the table "groups" should be:
      | id  | type     | name       |
      | 1   | Base     | Root       |
      | 2   | Base     | RootSelf   |
      | 4   | Base     | RootTemp   |
      | 31  | UserSelf | tmp-1234   |
      | 100 | Class    | Some class |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id |
      | 1               | 2              |
      | 2               | 4              |
      | 4               | 31             |
    And the table "group_pending_requests" should be:
      | group_id | member_id |
      | 100      | 31        |
    And the table "group_membership_changes" should be:
      | group_id | member_id |
      | 100      | 31        |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 1                 | 1              | true    |
      | 1                 | 2              | false   |
      | 1                 | 4              | false   |
      | 1                 | 31             | false   |
      | 2                 | 2              | true    |
      | 2                 | 4              | false   |
      | 2                 | 31             | false   |
      | 4                 | 4              | true    |
      | 4                 | 31             | false   |
      | 31                | 31             | true    |
      | 100               | 100            | true    |

  Scenario: Temporary user
    Given I am the user with id "31"
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
      | temp_user | login | group_id |
      | 0         | user  | 21       |
    And the table "groups" should be:
      | id  | type      | name       |
      | 1   | Base      | Root       |
      | 2   | Base      | RootSelf   |
      | 4   | Base      | RootTemp   |
      | 21  | UserSelf  | user       |
      | 100 | Class     | Some class |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id |
      | 1               | 2              |
      | 2               | 4              |
      | 2               | 21             |
    And the table "group_pending_requests" should be:
      | group_id | member_id |
      | 100      | 21        |
    And the table "group_membership_changes" should be:
      | group_id | member_id |
      | 100      | 21        |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 1                 | 1              | true    |
      | 1                 | 2              | false   |
      | 1                 | 4              | false   |
      | 1                 | 21             | false   |
      | 2                 | 2              | true    |
      | 2                 | 4              | false   |
      | 2                 | 21             | false   |
      | 4                 | 4              | true    |
      | 21                | 21             | true    |
      | 100               | 100            | true    |
