Feature: Create a user batch
  Background:
    Given the database has the following table "groups":
      | id | type    | name     | created_at          | require_personal_info_access_approval | require_lock_membership_approval_until | require_watch_approval |
      | 2  | Base    | AllUsers | 2015-08-10 12:34:55 | none                                  | null                                   | 0                      |
      | 3  | Club    | Club     | 2017-08-10 12:34:55 | view                                  | 3030-01-01 00:00:00                    | 1                      |
      | 4  | Friends | Friends  | 2018-08-10 12:34:55 | edit                                  | 2019-01-01 00:00:00                    | 0                      |
      | 21 | User    | owner    | 2016-08-10 12:34:55 | none                                  | null                                   | 0                      |
    And the database has the following table "users":
      | login | group_id | first_name  | last_name | default_language |
      | owner | 21       | Jean-Michel | Blanquer  | en               |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage            |
      | 3        | 21         | memberships           |
      | 4        | 21         | memberships_and_group |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 2               | 21             |
      | 3               | 4              |
    And the groups ancestors are computed
    And the database has the following table "user_batch_prefixes":
      | group_prefix | group_id | allow_new | max_users |
      | test         | 3        | 1         | 2         |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: 2
      """

  Scenario: Create a new user batch
    Given the time now is "2019-07-17T01:02:29+03:00"
    And the DB time now is "2019-07-16 22:02:28"
    And the login module "create" endpoint with params "amount=2&language=en&login_fixed=1&password_length=6&postfix_length=3&prefix=test_custom_" returns 200 with encoded body:
      """
      {
        "success": true,
        "data": [
          {"id":100000029,"login":"test_custom_jzk","password":"fy52ka"},
          {"id":100000030,"login":"test_custom_ctc","password":"aa3k7i"}
        ]
      }
      """
    And I am the user with id "21"
    When I send a POST request to "/user-batches" with the following body:
      """
      {
        "custom_prefix":"custom",
        "group_prefix":"test",
        "password_length":6,
        "postfix_length":3,
        "subgroups":[
          {"count":1,"group_id":4},
          {"count":1,"group_id":3}
        ]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": [
          {
            "group_id": "4",
            "users": [
              {
                "login": "test_custom_jzk",
                "password": "fy52ka",
                "user_id": "5577006791947779410"
              }
            ]
          },
          {
            "group_id": "3",
            "users": [
              {
                "login": "test_custom_ctc",
                "password": "aa3k7i",
                "user_id": "8674665223082153551"
              }
            ]
          }
        ]
      }
      """
    And the table "user_batches_v2" should be:
      | group_prefix | custom_prefix | size | creator_id | created_at          |
      | test         | custom        | 2    | 21         | 2019-07-16 22:02:28 |
    And the table "users" should be:
      | group_id            | latest_login_at | latest_activity_at | temp_user | registered_at       | login_id  | login           | default_language | email | first_name  | last_name |
      | 21                  | null            | null               | 0         | null                | null      | owner           | en               | null  | Jean-Michel | Blanquer  |
      | 5577006791947779410 | null            | null               | 0         | 2019-07-16 22:02:28 | 100000029 | test_custom_jzk | en               | null  | null        | null      |
      | 8674665223082153551 | null            | null               | 0         | 2019-07-16 22:02:28 | 100000030 | test_custom_ctc | en               | null  | null        | null      |
    And the table "groups" should be:
      | id                  | name            | type    | description     | created_at          | is_open | send_emails |
      | 2                   | AllUsers        | Base    | null            | 2015-08-10 12:34:55 | false   | false       |
      | 3                   | Club            | Club    | null            | 2017-08-10 12:34:55 | false   | false       |
      | 4                   | Friends         | Friends | null            | 2018-08-10 12:34:55 | false   | false       |
      | 21                  | owner           | User    | null            | 2016-08-10 12:34:55 | false   | false       |
      | 5577006791947779410 | test_custom_jzk | User    | test_custom_jzk | 2019-07-16 22:02:28 | false   | false       |
      | 8674665223082153551 | test_custom_ctc | User    | test_custom_ctc | 2019-07-16 22:02:28 | false   | false       |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id      | personal_info_view_approved_at | lock_membership_approved_at | watch_approved_at   |
      | 2               | 21                  | null                           | null                        | null                |
      | 2               | 5577006791947779410 | null                           | null                        | null                |
      | 2               | 8674665223082153551 | null                           | null                        | null                |
      | 3               | 4                   | null                           | null                        | null                |
      | 3               | 8674665223082153551 | 2019-07-16 22:02:28            | 2019-07-16 22:02:28         | 2019-07-16 22:02:28 |
      | 4               | 5577006791947779410 | 2019-07-16 22:02:28            | null                        | null                |
    And the table "groups_ancestors" should be:
      | ancestor_group_id   | child_group_id      | is_self |
      | 2                   | 2                   | true    |
      | 2                   | 21                  | false   |
      | 2                   | 5577006791947779410 | false   |
      | 2                   | 8674665223082153551 | false   |
      | 3                   | 3                   | true    |
      | 3                   | 4                   | false   |
      | 3                   | 5577006791947779410 | false   |
      | 3                   | 8674665223082153551 | false   |
      | 4                   | 4                   | true    |
      | 4                   | 5577006791947779410 | false   |
      | 21                  | 21                  | true    |
      | 5577006791947779410 | 5577006791947779410 | true    |
      | 8674665223082153551 | 8674665223082153551 | true    |
    And the table "attempts" should be:
      | participant_id      | id | creator_id          | ABS(TIMESTAMPDIFF(SECOND, NOW(), created_at)) < 3 | parent_attempt_id | root_item_id |
      | 5577006791947779410 | 0  | 5577006791947779410 | true                                              | null              | null         |
      | 8674665223082153551 | 0  | 8674665223082153551 | true                                              | null              | null         |
    And the table "group_membership_changes" should be empty
