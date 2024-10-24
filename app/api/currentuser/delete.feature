Feature: Delete the current user
  Background:
    Given the database has the following table "groups":
      | id  | type  | name       | require_lock_membership_approval_until |
      | 2   | Base  | AllUsers   | 9999-12-31 23:59:59                    |
      | 4   | Base  | TempUsers  | null                                   |
      | 21  | User  | user       | null                                   |
      | 31  | User  | tmp-1234   | null                                   |
      | 100 | Class | Some class | null                                   |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 2               | 4              |
      | 2               | 21             |
      | 4               | 31             |
    And the groups ancestors are computed
    And the database has the following table "group_pending_requests":
      | group_id | member_id | at                  |
      | 100      | 21        | 2019-05-30 11:00:00 |
      | 100      | 31        | 2019-05-30 11:00:00 |
    And the database has the following table "group_membership_changes":
      | group_id | member_id | at                      |
      | 100      | 21        | 2019-05-30 11:00:00.001 |
      | 100      | 31        | 2019-05-30 11:00:00.001 |
    And the database has the following users:
      | group_id | temp_user | login    | login_id |
      | 21       | 0         | user     | 1234567  |
      | 31       | 1         | tmp-1234 | null     |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
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
      | id  | type  | name       |
      | 2   | Base  | AllUsers   |
      | 4   | Base  | TempUsers  |
      | 31  | User  | tmp-1234   |
      | 100 | Class | Some class |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id |
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
      | id  | type  | name       |
      | 2   | Base  | AllUsers   |
      | 4   | Base  | TempUsers  |
      | 21  | User  | user       |
      | 100 | Class | Some class |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id |
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
      | 2                 | 2              | true    |
      | 2                 | 4              | false   |
      | 2                 | 21             | false   |
      | 4                 | 4              | true    |
      | 21                | 21             | true    |
      | 100               | 100            | true    |
