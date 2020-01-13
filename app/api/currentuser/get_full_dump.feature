Feature: Export the current user's data
  Background:
    Given the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'groups':
      | id | type     | name               | description            |
      | 1  | Class    | Our Class          | Our class group        |
      | 2  | Team     | Our Team           | Our team group         |
      | 3  | Club     | Our Club           | Our club group         |
      | 4  | Friends  | Our Friends        | Group for our friends  |
      | 5  | Other    | Other people       | Group for other people |
      | 6  | Class    | Another Class      | Another class group    |
      | 7  | Team     | Another Team       | Another team group     |
      | 8  | Club     | Another Club       | Another club group     |
      | 9  | Friends  | Some other friends | Another friends group  |
      | 10 | Other    | Secret group       | Secret group           |
      | 11 | UserSelf | user self          |                        |
      | 21 | UserSelf | jack               |                        |
      | 31 | UserSelf | jane               |                        |
    And the database has the following table 'users':
      | login | group_id | first_name | last_name | grade |
      | user  | 11       | John       | Doe       | 1     |
      | jack  | 21       | Jack       | Smith     | 2     |
      | jane  | 31       | Jane       | Doe       | 2     |
    And the database has the following table 'refresh_tokens':
      | user_id | refresh_token    |
      | 21      | refreshTokenFor1 |
      | 11      | refreshTokenFor2 |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 3  | 2               | 11             |
      | 6  | 5               | 11             |
      | 7  | 6               | 11             |
      | 10 | 9               | 11             |
      | 14 | 10              | 11             |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            | can_grant_group_access | can_watch_members |
      | 1        | 11         | memberships           | 1                      | 0                 |
      | 2        | 11         | memberships_and_group | 0                      | 1                 |
      | 6        | 9          | memberships           | 1                      | 1                 |
      | 9        | 8          | none                  | 0                      | 0                 |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         | at                  |
      | 1        | 11        | invitation   | 2019-08-10 00:00:00 |
      | 3        | 11        | join_request | 2019-08-11 00:00:00 |
      | 1        | 21        | invitation   | 2019-08-12 00:00:00 |
    And the database has the following table 'group_membership_changes':
      | group_id | member_id | action               | at                  | initiator_id |
      | 4        | 11        | join_request_refused | 2019-07-10 00:02:28 | 11           |
      | 7        | 11        | removed              | 2019-07-10 03:02:28 | 31           |
      | 8        | 11        | left                 | 2019-07-10 04:02:28 | 11           |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 1                 | 1              | true    |
      | 2                 | 2              | true    |
      | 2                 | 11             | false   |
      | 3                 | 3              | true    |
      | 4                 | 4              | true    |
      | 5                 | 5              | true    |
      | 5                 | 11             | false   |
      | 6                 | 6              | true    |
      | 6                 | 11             | false   |
      | 7                 | 7              | true    |
      | 8                 | 8              | true    |
      | 9                 | 9              | true    |
      | 9                 | 11             | false   |
      | 10                | 11             | false   |
      | 11                | 11             | true    |
    And the database has the following table 'items':
      | id  |
      | 404 |
      | 405 |
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 111 | 11       | 404     | 0     |
      | 112 | 2        | 404     | 0     |
      | 113 | 1        | 405     | 0     |
    And the database has the following table 'answers':
      | id | author_id | attempt_id | created_at          |
      | 1  | 11        | 111        | 2019-07-09 21:02:28 |
      | 2  | 21        | 113        | 2019-07-09 21:02:28 |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 11      | 404     | 111               |
      | 21      | 405     | 112               |

  Scenario: Full data
    Given I am the user with id "11"
    When I send a GET request to "/current-user/full-dump"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "attachment; filename=user_data.json"
    And the response body should be, in JSON:
    """
    {
      "current_user": {
        "group_id": "11", "allow_subgroups": null, "basic_editor_mode": 1, "email_verified": 0, "is_admin": 0,
        "no_ranking": 0, "notify_news": 0, "photo_autoload": 0, "public_first_name": 0, "public_last_name": 0,
        "creator_id": null, "grade": 1, "graduation_year": 0, "member_state": 0, "step_level_in_site": 0,
        "access_group_id": null, "login_id": null,
        "login_module_prefix": null, "help_given": 0, "spaces_for_tab": 3, "address": null, "birth_date": null,
        "cell_phone_number": null, "city": null, "country_code": "", "default_language": "fr", "email": null,
        "first_name": "John", "free_text": null, "land_line_number": null, "lang_prog": "Python",
        "latest_activity_at": null, "last_ip": null, "latest_login_at": null, "last_name": "Doe", "login": "user",
        "notifications_read_at": null, "notify": "Answers", "open_id_identity": null, "password_md5": null,
        "recover": null, "registered_at": null, "salt": null, "sex": null, "student_id": null, "time_zone": null,
        "web_site": null, "zipcode": null, "temp_user": 0
      },
      "groups_attempts": [
        {
          "id": "111", "finished": 0, "validated": 0,
          "order": 0, "score_computed": 0, "score_edit_rule": null, "score_edit_value": null,
          "group_id": "11", "item_id": "404", "creator_id": null, "children_validated": 0,
          "hints_cached": 0, "submissions": 0, "tasks_solved": 0, "tasks_tried": 0,
          "tasks_with_help": 0, "result_propagation_state": "done",
          "score_obtained_at": null, "entered_at": null, "hints_requested": null,
          "latest_activity_at": null, "latest_answer_at": null, "latest_hint_at": null, "score_edit_comment": null,
          "started_at": null, "validated_at": null
        },
        {
          "id": "112", "finished": 0, "validated": 0,
          "order": 0, "score_computed": 0, "score_edit_rule": null, "score_edit_value": null,
          "group_id": "2", "item_id": "404", "creator_id": null, "children_validated": 0,
          "hints_cached": 0, "submissions": 0, "tasks_solved": 0, "tasks_tried": 0,
          "tasks_with_help": 0, "result_propagation_state": "done",
          "score_obtained_at": null, "entered_at": null, "hints_requested": null,
          "latest_activity_at": null, "latest_answer_at": null, "latest_hint_at": null, "score_edit_comment": null,
          "started_at": null, "validated_at": null
        }
      ],
      "groups_groups": [
        {
          "id": "3", "child_order": 0, "child_group_id": "11", "parent_group_id": "2",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Our Team", "expires_at": "9999-12-31T23:59:59Z"
        },
        {
          "id": "6", "child_order": 0, "child_group_id": "11", "parent_group_id": "5",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Other people", "expires_at": "9999-12-31T23:59:59Z"
        },
        {
          "id": "7", "child_order": 0, "child_group_id": "11", "parent_group_id": "6",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Another Class", "expires_at": "9999-12-31T23:59:59Z"
        },
        {
          "id": "10", "child_order": 0, "child_group_id": "11", "parent_group_id": "9",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Some other friends", "expires_at": "9999-12-31T23:59:59Z"
        },
        {
          "id": "14", "child_order": 0, "child_group_id": "11", "parent_group_id": "10",
          "lock_membership_approved_at": null, "lock_membership_approved": 0,
          "personal_info_view_approved_at": null, "personal_info_view_approved": 0,
          "watch_approved_at": null, "watch_approved": 0,
          "name": "Secret group", "expires_at": "9999-12-31T23:59:59Z"
        }
      ],
      "group_managers": [
        {
          "can_grant_group_access": 1,
          "can_manage": "memberships",
          "can_watch_members": 0,
          "can_edit_personal_info": 0,
          "group_id": "1",
          "manager_id": "11",
          "name": "Our Class"
        },
        {
          "can_grant_group_access": 0,
          "can_manage": "memberships_and_group",
          "can_watch_members": 1,
          "can_edit_personal_info": 0,
          "group_id": "2",
          "manager_id": "11",
          "name": "Our Team"
        }
      ],
      "group_membership_changes": [
        {
          "action": "left", "at": "2019-07-10T04:02:28Z", "group_id": "8", "initiator_id": "11",
          "member_id": "11", "name": "Another Club"
        },
        {
          "action": "removed", "at": "2019-07-10T03:02:28Z", "group_id": "7", "initiator_id": "31",
          "member_id": "11", "name": "Another Team"
        },
        {
          "action": "join_request_refused", "at": "2019-07-10T00:02:28Z", "group_id": "4", "initiator_id": "11",
          "member_id": "11", "name": "Our Friends"
        }
      ],
      "group_pending_requests": [
        {
          "group_id": "1", "member_id": "11", "name": "Our Class", "type": "invitation",
          "lock_membership_approved": 0, "personal_info_view_approved": 0,
          "watch_approved": 0, "at": "2019-08-10T00:00:00Z"
        },
        {
          "group_id": "3", "member_id": "11", "name": "Our Club", "type": "join_request",
          "lock_membership_approved": 0, "personal_info_view_approved": 0,
          "watch_approved": 0, "at": "2019-08-11T00:00:00Z"
        }
      ],
      "joined_groups": [
        {"id": "2", "name": "Our Team"},
        {"id": "5", "name": "Other people"},
        {"id": "6", "name": "Another Class"},
        {"id": "9", "name": "Some other friends"},
        {"id": "10", "name": "Secret group"}
      ],
      "managed_groups": [
        {"id": "1", "name": "Our Class"},
        {"id": "2", "name": "Our Team"},
        {"id": "6", "name": "Another Class"},
        {"id": "11", "name": "user self"}
      ],
      "refresh_token": {"user_id": "11", "refresh_token": "***"},
      "sessions": [
        {
          "user_id": "11", "access_token": "***", "expires_at": "2019-07-17T00:02:28Z",
          "issued_at": "2019-07-16T22:02:28Z", "issuer": null
        }
      ],
      "answers": [
        {
          "id": "1", "score": null, "attempt_id": "111",
          "author_id": "11", "answer": null, "graded_at": null,
          "state": null, "created_at": "2019-07-09T21:02:28Z", "type": "Submission"
        }
      ],
      "users_items": [
        {
          "active_attempt_id": "111", "item_id": "404", "user_id": "11"
        }
      ]
    }
    """

  Scenario: All optional arrays and objects are empty
    Given I am the user with id "31"
    When I send a GET request to "/current-user/full-dump"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "attachment; filename=user_data.json"
    And the response body should be, in JSON:
    """
    {
      "current_user": {
        "group_id": "31", "allow_subgroups": null, "basic_editor_mode": 1, "email_verified": 0, "is_admin": 0, "no_ranking": 0,
        "notify_news": 0, "photo_autoload": 0, "public_first_name": 0, "public_last_name": 0, "creator_id": null,
        "grade": 2, "graduation_year": 0, "member_state": 0, "step_level_in_site": 0, "access_group_id": null,
        "login_id": null, "login_module_prefix": null,
        "help_given": 0, "spaces_for_tab": 3, "address": null, "birth_date": null, "cell_phone_number": null,
        "city": null, "country_code": "", "default_language": "fr", "email": null, "first_name": "Jane",
        "free_text": null, "land_line_number": null, "lang_prog": "Python", "latest_activity_at": null, "last_ip": null,
        "latest_login_at": null, "last_name": "Doe", "login": "jane", "notifications_read_at": null, "notify": "Answers",
        "open_id_identity": null, "password_md5": null, "recover": null, "registered_at": null, "salt": null,
        "sex": null, "student_id": null, "time_zone": null, "web_site": null, "zipcode": null, "temp_user": 0
      },
      "groups_attempts": [],
      "groups_groups": [],
      "group_managers": [],
      "group_membership_changes": [],
      "group_pending_requests": [],
      "joined_groups": [],
      "managed_groups": [],
      "refresh_token": null,
      "sessions": [
        {
          "user_id": "31", "access_token": "***", "expires_at": "2019-07-17T00:02:28Z",
          "issued_at": "2019-07-16T22:02:28Z", "issuer": null
        }
      ],
      "answers": [],
      "users_items": []
    }
    """
