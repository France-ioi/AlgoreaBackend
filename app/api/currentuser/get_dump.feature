Feature: Export the short version of the current user's data
  Background:
    Given the DB time now is "2019-07-16 22:02:28"
    And the database has the following table "groups":
      | id | type    | name               | description            |
      | 1  | Class   | Our Class          | Our class group        |
      | 2  | Team    | Our Team           | Our team group         |
      | 3  | Club    | Our Club           | Our club group         |
      | 4  | Friends | Our Friends        | Group for our friends  |
      | 5  | Other   | Other people       | Group for other people |
      | 6  | Class   | Another Class      | Another class group    |
      | 7  | Team    | Another Team       | Another team group     |
      | 8  | Club    | Another Club       | Another club group     |
      | 9  | Friends | Some other friends | Another friends group  |
      | 11 | User    | user self          |                        |
      | 31 | User    | jane               |                        |
    And the database has the following users:
      | group_id | login | first_name | last_name | grade |
      | 11       | user  | John       | Doe       | 1     |
      | 31       | jane  | Jane       | Doe       | 2     |
    And the database has the following table "sessions":
      | session_id | user_id | refresh_token    |
      | 1          | 11      | refreshTokenFor1 |
      | 2          | 31      | refreshTokenFor2 |
    And the database has the following table "access_tokens":
      | session_id | token        | expires_at          |
      | 1          | accessToken1 | 3000-01-01 00:00:00 |
      | 2          | accessToken2 | 3000-01-01 00:00:00 |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 2               | 11             |
      | 5               | 11             |
      | 6               | 11             |
      | 9               | 11             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage            | can_grant_group_access | can_watch_members |
      | 1        | 11         | memberships           | 1                      | 0                 |
      | 2        | 11         | memberships_and_group | 0                      | 1                 |
      | 6        | 9          | memberships           | 1                      | 1                 |
      | 9        | 8          | none                  | 0                      | 0                 |
    And the database has the following table "group_pending_requests":
      | group_id | member_id | type         |
      | 1        | 11        | invitation   |
      | 3        | 11        | join_request |
      | 1        | 31        | invitation   |
    And the database has the following table "group_membership_changes":
      | group_id | member_id | action               | at                      | initiator_id |
      | 4        | 11        | join_request_refused | 2019-07-10 00:02:28.001 | 11           |
      | 7        | 11        | removed              | 2019-07-10 03:02:28.002 | 31           |
      | 8        | 11        | left                 | 2019-07-10 04:02:28.003 | 11           |
    And the database has the following table "items":
      | id  | default_language_tag |
      | 404 | fr                   |
    And the database has the following table "attempts":
      | id | participant_id |
      | 0  | 11             |
      | 0  | 2              |
      | 0  | 1              |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id |
      | 0          | 11             | 404     |
      | 0          | 2              | 404     |
      | 0          | 1              | 404     |
    And the database has the following table "answers":
      | id | author_id | participant_id | attempt_id | item_id | created_at          |
      | 1  | 11        | 11             | 0          | 404     | 2019-07-09 20:02:28 |
      | 2  | 31        | 1              | 0          | 404     | 2019-07-09 20:02:28 |

  Scenario: Full data
    Given I am the user with id "11"
    When I send a GET request to "/current-user/dump"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "attachment; filename=user_data.json"
    And the response body should be, in JSON:
    """
    {
      "current_user": {
        "group_id": "11", "basic_editor_mode": 1, "email_verified": 0, "is_admin": 0,
        "no_ranking": 0, "notify_news": 0, "photo_autoload": 0, "public_first_name": 0, "public_last_name": 0,
        "creator_id": null, "grade": 1, "graduation_year": 0, "member_state": 0, "step_level_in_site": 0,
        "access_group_id": null, "login_id": null, "latest_profile_sync_at": null,
        "help_given": 0, "spaces_for_tab": 3, "address": null, "birth_date": null,
        "cell_phone_number": null, "city": null, "country_code": "", "default_language": "fr", "email": null,
        "first_name": "John", "free_text": null, "land_line_number": null, "lang_prog": "Python",
        "latest_activity_at": null, "last_ip": null, "latest_login_at": null, "last_name": "Doe", "login": "user",
        "notifications_read_at": null, "notify": "Answers", "open_id_identity": null, "password_md5": null,
        "recover": null, "registered_at": null, "salt": null, "sex": null, "student_id": null, "time_zone": null,
        "web_site": null, "zipcode": null, "temp_user": 0
      },
      "groups_groups": [
        {
          "child_group_id": "11", "parent_group_id": "2",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Our Team", "expires_at": "9999-12-31T23:59:59Z",
          "is_team_membership": 1
        },
        {
          "child_group_id": "11", "parent_group_id": "5",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Other people", "expires_at": "9999-12-31T23:59:59Z",
          "is_team_membership": 0
        },
        {
          "child_group_id": "11", "parent_group_id": "6",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Another Class", "expires_at": "9999-12-31T23:59:59Z",
          "is_team_membership": 0
        },
        {
          "child_group_id": "11", "parent_group_id": "9",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Some other friends", "expires_at": "9999-12-31T23:59:59Z",
          "is_team_membership": 0
        }
      ],
      "group_managers": [
        {
          "can_grant_group_access": 1,
          "can_manage": "memberships",
          "can_manage_value": 2,
          "can_watch_members": 0,
          "can_edit_personal_info": 0,
          "group_id": "1",
          "manager_id": "11",
          "name": "Our Class"
        },
        {
          "can_grant_group_access": 0,
          "can_manage": "memberships_and_group",
          "can_manage_value": 3,
          "can_watch_members": 1,
          "can_edit_personal_info": 0,
          "group_id": "2",
          "manager_id": "11",
          "name": "Our Team"
        }
      ],
      "joined_groups": [
        {"id": "2", "name": "Our Team"},
        {"id": "5", "name": "Other people"},
        {"id": "6", "name": "Another Class"},
        {"id": "9", "name": "Some other friends"}
      ],
      "managed_groups": [
        {"id": "1", "name": "Our Class"},
        {"id": "2", "name": "Our Team"},
        {"id": "6", "name": "Another Class"},
        {"id": "11", "name": "user self"}
      ]
    }
    """

  Scenario: All optional arrays and objects are empty
    Given I am the user with id "31"
    When I send a GET request to "/current-user/dump"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "attachment; filename=user_data.json"
    And the response body should be, in JSON:
    """
    {
      "current_user": {
        "group_id": "31", "basic_editor_mode": 1, "email_verified": 0, "is_admin": 0, "no_ranking": 0,
        "notify_news": 0, "photo_autoload": 0, "public_first_name": 0, "public_last_name": 0, "creator_id": null,
        "grade": 2, "graduation_year": 0, "member_state": 0, "step_level_in_site": 0, "access_group_id": null,
        "login_id": null, "help_given": 0, "spaces_for_tab": 3, "address": null, "birth_date": null, "cell_phone_number": null,
        "city": null, "country_code": "", "default_language": "fr", "email": null, "first_name": "Jane",
        "free_text": null, "land_line_number": null, "lang_prog": "Python", "latest_activity_at": null, "last_ip": null,
        "latest_login_at": null, "last_name": "Doe", "login": "jane", "notifications_read_at": null, "notify": "Answers",
        "open_id_identity": null, "password_md5": null, "recover": null, "registered_at": null, "salt": null,
        "sex": null, "student_id": null, "time_zone": null, "web_site": null, "zipcode": null, "temp_user": 0,
        "latest_profile_sync_at": null
      },
      "groups_groups": [],
      "group_managers": [],
      "joined_groups": [],
      "managed_groups": []
    }
    """
