Feature: Export the current user's data
  Background:
    Given the DB time now is "2019-07-16 22:02:28"
    And the database has the following table 'users':
      | id | login | self_group_id | owned_group_id | first_name  | last_name | grade |
      | 2  | user  | 11            | 12             | John        | Doe       | 1     |
      | 4  | jane  | 31            | 32             | Jane        | Doe       | 2     |
    And the database has the following table 'refresh_tokens':
      | user_id | refresh_token    |
      | 1       | refreshTokenFor1 |
      | 2       | refreshTokenFor2 |
    And the database has the following table 'groups':
      | id | type      | name               | description            |
      | 1  | Class     | Our Class          | Our class group        |
      | 2  | Team      | Our Team           | Our team group         |
      | 3  | Club      | Our Club           | Our club group         |
      | 4  | Friends   | Our Friends        | Group for our friends  |
      | 5  | Other     | Other people       | Group for other people |
      | 6  | Class     | Another Class      | Another class group    |
      | 7  | Team      | Another Team       | Another team group     |
      | 8  | Club      | Another Club       | Another club group     |
      | 9  | Friends   | Some other friends | Another friends group  |
      | 10 | Other     | Secret group       | Secret group           |
      | 11 | UserSelf  | user self          |                        |
      | 12 | UserAdmin | user admin         |                        |
      | 31 | UserSelf  | jane               |                        |
      | 32 | UserAdmin | jane-admin         |                        |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type               | status_changed_at   | inviting_user_id |
      | 2  | 1               | 11             | invitationSent     | 2019-07-09 21:02:28 | null             |
      | 3  | 2               | 11             | invitationAccepted | 2019-07-09 22:02:28 | 1                |
      | 4  | 3               | 11             | requestSent        | 2019-07-09 23:02:28 | 1                |
      | 5  | 4               | 11             | requestRefused     | 2019-07-10 00:02:28 | 2                |
      | 6  | 5               | 11             | invitationAccepted | 2019-07-10 01:02:28 | 2                |
      | 7  | 6               | 11             | requestAccepted    | 2019-07-10 02:02:28 | 2                |
      | 8  | 7               | 11             | removed            | 2019-07-10 03:02:28 | 1                |
      | 9  | 8               | 11             | left               | 2019-07-10 04:02:28 | 1                |
      | 10 | 9               | 11             | direct             | 2019-07-10 05:02:28 | 2                |
      | 11 | 1               | 12             | invitationSent     | 2019-07-09 20:02:28 | 2                |
      | 12 | 12              | 1              | direct             | null                | null             |
      | 13 | 12              | 2              | direct             | null                | null             |
      | 14 | 10              | 11             | joinedByCode       | 2019-07-10 05:02:28 | null             |
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
      | 12                | 1              | false   |
      | 12                | 2              | false   |
      | 12                | 12             | true    |
    And the database has the following table 'users_answers':
      | id | user_id | item_id | submitted_at        |
      | 1  | 2       | 404     | 2019-07-09 21:02:28 |
      | 2  | 3       | 405     | 2019-07-09 21:02:28 |
    And the database has the following table 'users_items':
      | id | user_id | item_id |
      | 11 | 2       | 404     |
      | 12 | 3       | 405     |
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 111 | 11       | 404     | 0     |
      | 112 | 2        | 404     | 0     |
      | 113 | 1        | 405     | 0     |

  Scenario: Full data
    Given I am the user with id "2"
    When I send a GET request to "/current-user/full-dump"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "attachment; filename=user_data.json"
    And the response body should be, in JSON:
    """
    {
      "current_user": {
        "id": "2", "allow_subgroups": null, "basic_editor_mode": 1, "email_verified": 0, "is_admin": 0,
        "no_ranking": 0, "notify_news": 0, "photo_autoload": 0, "public_first_name": 0, "public_last_name": 0,
        "creator_id": null, "grade": 1, "graduation_year": 0, "member_state": 0, "step_level_in_site": 0,
        "access_group_id": null, "owned_group_id": "12", "self_group_id": "11", "godfather_user_id": null, "login_id": null,
        "login_module_prefix": null, "help_given": 0, "spaces_for_tab": 3, "address": null, "birth_date": null,
        "cell_phone_number": null, "city": null, "country_code": "", "default_language": "fr", "email": null,
        "first_name": "John", "free_text": null, "land_line_number": null, "lang_prog": "Python",
        "last_activity_at": null, "last_ip": null, "last_login_at": null, "last_name": "Doe", "login": "user",
        "notifications_read_at": null, "notify": "Answers", "open_id_identity": null, "password_md5": null,
        "recover": null, "registered_at": null, "salt": null, "sex": null, "student_id": null, "time_zone": null,
        "web_site": null, "zipcode": null, "temp_user": 0
      },
      "groups_attempts": [
        {
          "id": "111", "finished": 0, "key_obtained": 0, "ranked": 0, "validated": 0, "autonomy": 0, "minus_score": -0,
          "order": 0, "precision": 0, "score": 0, "score_computed": 0, "score_diff_manual": 0, "score_reeval": 0,
          "group_id": "11", "item_id": "404", "creator_user_id": null, "children_validated": 0,
          "corrections_read": 0, "hints_cached": 0, "submissions_attempts": 0, "tasks_solved": 0, "tasks_tried": 0,
          "tasks_with_help": 0, "all_lang_prog": null, "ancestors_computation_state": "done",
          "best_answer_at": null, "contest_started_at": null, "finished_at": null, "hints_requested": null,
          "last_activity_at": null, "last_answer_at": null, "last_hint_at": null, "score_diff_comment": "",
          "started_at": null, "thread_started_at": null, "validated_at": null
        },
        {
          "id": "112", "finished": 0, "key_obtained": 0, "ranked": 0, "validated": 0, "autonomy": 0, "minus_score": -0,
          "order": 0, "precision": 0, "score": 0, "score_computed": 0, "score_diff_manual": 0, "score_reeval": 0,
          "group_id": "2", "item_id": "404", "creator_user_id": null, "children_validated": 0,
          "corrections_read": 0, "hints_cached": 0, "submissions_attempts": 0, "tasks_solved": 0, "tasks_tried": 0,
          "tasks_with_help": 0, "all_lang_prog": null, "ancestors_computation_state": "done",
          "best_answer_at": null, "contest_started_at": null, "finished_at": null, "hints_requested": null,
          "last_activity_at": null, "last_answer_at": null, "last_hint_at": null, "score_diff_comment": "",
          "started_at": null, "thread_started_at": null, "validated_at": null
        }
      ],
      "groups_groups": [
        {
          "id": "2", "child_order": 0, "child_group_id": "11", "parent_group_id": "1", "inviting_user_id": null,
          "name": "Our Class", "role": "member", "status_changed_at": "2019-07-09T21:02:28Z", "type": "invitationSent"
        },
        {
          "id": "3", "child_order": 0, "child_group_id": "11", "parent_group_id": "2", "inviting_user_id": "1",
          "name": "Our Team", "role": "member", "status_changed_at": "2019-07-09T22:02:28Z", "type": "invitationAccepted"
        },
        {
          "id": "4", "child_order": 0, "child_group_id": "11", "parent_group_id": "3", "inviting_user_id": "1",
          "name": "Our Club", "role": "member", "status_changed_at": "2019-07-09T23:02:28Z", "type": "requestSent"
        },
        {
          "id": "5", "child_order": 0, "child_group_id": "11", "parent_group_id": "4", "inviting_user_id": "2",
          "name": "Our Friends", "role": "member", "status_changed_at": "2019-07-10T00:02:28Z", "type": "requestRefused"
        },
        {
          "id": "6", "child_order": 0, "child_group_id": "11", "parent_group_id": "5", "inviting_user_id": "2",
          "name": "Other people", "role": "member", "status_changed_at": "2019-07-10T01:02:28Z", "type": "invitationAccepted"
        },
        {
          "id": "7", "child_order": 0, "child_group_id": "11", "parent_group_id": "6", "inviting_user_id": "2",
          "name": "Another Class", "role": "member", "status_changed_at": "2019-07-10T02:02:28Z", "type": "requestAccepted"
        },
        {
          "id": "8", "child_order": 0, "child_group_id": "11", "parent_group_id": "7", "inviting_user_id": "1",
          "name": "Another Team", "role": "member", "status_changed_at": "2019-07-10T03:02:28Z", "type": "removed"
        },
        {
          "id": "9", "child_order": 0, "child_group_id": "11", "parent_group_id": "8", "inviting_user_id": "1",
          "name": "Another Club", "role": "member", "status_changed_at": "2019-07-10T04:02:28Z", "type": "left"
        },
        {
          "id": "10", "child_order": 0, "child_group_id": "11", "parent_group_id": "9", "inviting_user_id": "2",
          "name": "Some other friends", "role": "member", "status_changed_at": "2019-07-10T05:02:28Z", "type": "direct"
        },
        {
          "id": "14", "child_order": 0, "child_group_id": "11", "parent_group_id": "10", "inviting_user_id": null,
          "name": "Secret group", "role": "member", "status_changed_at": "2019-07-10T05:02:28Z", "type": "joinedByCode"
        }
      ],
      "joined_groups": [
        {"id": "2", "name": "Our Team"},
        {"id": "5", "name": "Other people"},
        {"id": "6", "name": "Another Class"},
        {"id": "9", "name": "Some other friends"},
        {"id": "10", "name": "Secret group"}
      ],
      "owned_groups": [
        {"id": "1", "name": "Our Class"},
        {"id": "2", "name": "Our Team"}
      ],
      "refresh_token": {"user_id": "2", "refresh_token": "***"},
      "sessions": [
        {
          "user_id": "2", "access_token": "***", "expires_at": "2019-07-17T00:02:28Z",
          "issued_at": "2019-07-16T22:02:28Z", "issuer": null
        }
      ],
      "users_answers": [
        {
          "id": "1", "validated": null, "score": null, "attempt_id": null, "item_id": "404",
          "user_id": "2", "grader_user_id": null, "answer": null, "graded_at": null, "lang_prog": null,
          "name": null, "state": null, "submitted_at": "2019-07-09T21:02:28Z", "type": "Submission"
        }
      ],
      "users_items": [
        {
          "id": "11", "finished": 0, "key_obtained": 0, "platform_data_removed": 0, "ranked": 0, "validated": 0,
          "autonomy": 0, "precision": 0, "score": 0, "score_computed": 0, "score_diff_manual": 0, "score_reeval": 0,
          "active_attempt_id": null, "item_id": "404", "user_id": "2", "children_validated": 0,
          "corrections_read": 0, "hints_cached": 0, "submissions_attempts": 0, "tasks_solved": 0, "tasks_tried": 0,
          "tasks_with_help": 0, "all_lang_prog": null, "ancestors_computation_state": "todo",
          "answer": null, "best_answer_at": null, "contest_started_at": null, "finished_at": null,
          "hints_requested": null, "last_activity_at": null, "last_answer_at": null, "last_hint_at": null,
          "score_diff_comment": "", "started_at": null, "state": null, "thread_started_at": null, "validated_at": null
        }
      ]
    }
    """

  Scenario: All optional arrays and objects are empty
    Given I am the user with id "4"
    When I send a GET request to "/current-user/full-dump"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "attachment; filename=user_data.json"
    And the response body should be, in JSON:
    """
    {
      "current_user": {
        "id": "4", "allow_subgroups": null, "basic_editor_mode": 1, "email_verified": 0, "is_admin": 0, "no_ranking": 0,
        "notify_news": 0, "photo_autoload": 0, "public_first_name": 0, "public_last_name": 0, "creator_id": null,
        "grade": 2, "graduation_year": 0, "member_state": 0, "step_level_in_site": 0, "access_group_id": null,
        "owned_group_id": "32", "self_group_id": "31", "godfather_user_id": null, "login_id": null, "login_module_prefix": null,
        "help_given": 0, "spaces_for_tab": 3, "address": null, "birth_date": null, "cell_phone_number": null,
        "city": null, "country_code": "", "default_language": "fr", "email": null, "first_name": "Jane",
        "free_text": null, "land_line_number": null, "lang_prog": "Python", "last_activity_at": null, "last_ip": null,
        "last_login_at": null, "last_name": "Doe", "login": "jane", "notifications_read_at": null, "notify": "Answers",
        "open_id_identity": null, "password_md5": null, "recover": null, "registered_at": null, "salt": null,
        "sex": null, "student_id": null, "time_zone": null, "web_site": null, "zipcode": null, "temp_user": 0
      },
      "groups_attempts": [],
      "groups_groups": [],
      "joined_groups": [],
      "owned_groups": [],
      "refresh_token": null,
      "sessions": [
        {
          "user_id": "4", "access_token": "***", "expires_at": "2019-07-17T00:02:28Z",
          "issued_at": "2019-07-16T22:02:28Z", "issuer": null
        }
      ],
      "users_answers": [],
      "users_items": []
    }
    """
