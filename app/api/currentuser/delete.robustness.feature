Feature: Delete the current user - robustness
  Background:
    Given the DB time now is "2019-08-09 23:59:59"
    And the database has the following table 'groups':
      | id | type      | name       | lock_user_deletion_date |
      | 1  | Base      | Root       | null                    |
      | 2  | Base      | RootSelf   | null                    |
      | 3  | Base      | RootAdmin  | null                    |
      | 4  | Base      | RootTemp   | null                    |
      | 21 | UserSelf  | user       | null                    |
      | 22 | UserAdmin | user-admin | null                    |
      | 50 | Class     | Our class  | 2019-08-10              |
    And the database has the following table 'groups_groups':
      | group_parent_id | group_child_id | type               |
      | 1               | 2              | direct             |
      | 1               | 3              | direct             |
      | 1               | 50             | direct             |
      | 2               | 4              | direct             |
      | 2               | 21             | direct             |
      | 3               | 22             | direct             |
      | 50              | 21             | invitationAccepted |
    And the database has the following table 'groups_ancestors':
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 1              | true    |
      | 1                 | 2              | false   |
      | 1                 | 3              | false   |
      | 1                 | 4              | false   |
      | 1                 | 21             | false   |
      | 1                 | 22             | false   |
      | 1                 | 50             | false   |
      | 2                 | 2              | true    |
      | 2                 | 4              | false   |
      | 2                 | 21             | false   |
      | 3                 | 3              | true    |
      | 3                 | 22             | false   |
      | 4                 | 4              | true    |
      | 21                | 21             | true    |
      | 22                | 22             | true    |
      | 50                | 21             | false   |
      | 50                | 50             | true    |
    And the database has the following table 'users':
      | id | temp_user | login | group_self_id | group_owned_id | login_id |
      | 11 | 0         | user  | 21            | 22             | 1234567  |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
        callbackURL: "https://backend.algorea.org/auth/login-callback"
      """

  Scenario: User cannot delete himself right now
    Given I am the user with id "11"
    When I send a DELETE request to "/current-user"
    Then the response code should be 403
    And the response error message should contain "You cannot delete yourself right now"
    And logs should contain:
      """
      A user with id = 11 tried to delete himself, but he is a member of a group with lock_user_deletion_date >= NOW()
      """
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Login module fails
    Given I am the user with id "11"
    And the DB time now is "2019-08-10 00:00:00"
    And the login module "unlink_client" endpoint for user id "1234567" returns 500 with encoded body:
      """
      {"success":false}
      """
    When I send a DELETE request to "/current-user"
    Then the response code should be 500
    And the response error message should contain "Can't unlink the user"
    And the table "users" should be empty
    And the table "groups" should be:
      | id | type  | name      | lock_user_deletion_date |
      | 1  | Base  | Root      | null                    |
      | 2  | Base  | RootSelf  | null                    |
      | 3  | Base  | RootAdmin | null                    |
      | 4  | Base  | RootTemp  | null                    |
      | 50 | Class | Our class | 2019-08-10              |
    And the table "groups_groups" should be:
      | group_parent_id | group_child_id | type   |
      | 1               | 2              | direct |
      | 1               | 3              | direct |
      | 1               | 50             | direct |
      | 2               | 4              | direct |
    And the table "groups_ancestors" should be:
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 1              | true    |
      | 1                 | 2              | false   |
      | 1                 | 3              | false   |
      | 1                 | 4              | false   |
      | 1                 | 50             | false   |
      | 2                 | 2              | true    |
      | 2                 | 4              | false   |
      | 3                 | 3              | true    |
      | 4                 | 4              | true    |
      | 50                | 50             | true    |
