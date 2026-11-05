Feature: Thread activity log attempt order

  Background:
    Given the database has the following users:
      | login | group_id | default_language | profile                                    |
      | user  | 11       | en               | {"first_name": "John", "last_name": "Doe"} |
    And the database has the following table "groups":
      | id | type  | name      |
      | 13 | Class | Our Class |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 13              | 11             |
    And the groups ancestors are computed
    And the database has the following table "languages":
      | tag |
      | en  |
    And the database has the following table "items":
      | id  | type | no_score | default_language_tag | allows_multiple_attempts |
      | 500 | Task | false    | en                   | 1                        |
      | 501 | Task | false    | en                   | 0                        |
    And the database has the following table "items_strings":
      | item_id | language_tag | title  |
      | 500     | en           | Multi  |
      | 501     | en           | Single |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 11       | 500     | content            | answer              |
      | 11       | 501     | content            | answer              |
    And the database has the following table "attempts":
      | participant_id | id | parent_attempt_id | root_item_id |
      | 11             | 0  | null              | null         |
      | 11             | 1  | 0                 | 500          |
      | 11             | 2  | 0                 | 500          |
      | 11             | 4  | 0                 | 500          |
    And the database has the following table "results":
      | participant_id | attempt_id | item_id | started_at          | validated_at        |
      | 11             | 0          | 500     | 2019-05-28 10:00:00 | null                |
      | 11             | 1          | 500     | 2019-05-29 10:00:00 | 2019-05-29 14:00:00 |
      | 11             | 2          | 500     | 2019-05-30 10:00:00 | null                |
      | 11             | 4          | 500     | 2019-05-30 10:00:00 | null                |
      | 11             | 0          | 501     | 2019-04-01 10:00:00 | null                |
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | created_at          |
      | 100 | 11        | 11             | 0          | 500     | Submission | 2019-05-28 11:00:00 |
      | 101 | 11        | 11             | 1          | 500     | Submission | 2019-05-29 11:00:00 |
      | 102 | 11        | 11             | 1          | 500     | Saved      | 2019-05-29 12:00:00 |
      | 104 | 11        | 11             | 2          | 500     | Submission | 2019-05-30 11:00:00 |
      | 105 | 11        | 11             | 4          | 500     | Submission | 2019-05-30 11:00:00 |
      | 200 | 11        | 11             | 0          | 501     | Submission | 2019-04-01 11:00:00 |
    And the database has the following table "threads":
      | participant_id | item_id | helper_group_id | status | latest_update_at    |
      | 11             | 500     | 13              | closed | 2019-06-01 12:00:00 |
      | 11             | 501     | 13              | closed | 2019-04-01 12:00:00 |

  # Rows with a null validated_at still appear as result_validated; the thread log does not filter them out.
  Scenario: Thread log rows carry the same attempt order, including every activity of one result
    Given I am the user with id "11"
    When I send a GET request to "/items/500/participant/11/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission", "answer_id": "105", "at": "2019-05-30T11:00:00Z",
        "attempt_id": "4", "attempt_order": 4, "from_answer_id": "105", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "104", "at": "2019-05-30T11:00:00Z",
        "attempt_id": "2", "attempt_order": 3, "from_answer_id": "104", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-30T10:00:00Z",
        "attempt_id": "4", "attempt_order": 4, "from_answer_id": "104", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-30T10:00:00Z",
        "attempt_id": "2", "attempt_order": 3, "from_answer_id": "104", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_validated", "at": "2019-05-29T14:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "104", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "saved_answer", "answer_id": "102", "at": "2019-05-29T12:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "102", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "101", "at": "2019-05-29T11:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "101", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-29T10:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "101", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "100", "at": "2019-05-28T11:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "100", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-28T10:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "100", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_validated", "at": null,
        "attempt_id": "4", "attempt_order": 4, "from_answer_id": "100", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_validated", "at": null,
        "attempt_id": "2", "attempt_order": 3, "from_answer_id": "100", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_validated", "at": null,
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "100", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      }
    ]
    """

  Scenario: Thread log omits attempt_order on a single-attempt item
    Given I am the user with id "11"
    When I send a GET request to "/items/501/participant/11/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission", "answer_id": "200", "at": "2019-04-01T11:00:00Z",
        "attempt_id": "0", "from_answer_id": "200", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "501", "type": "Task", "string": {"title": "Single"}}
      },
      {
        "activity_type": "result_started", "at": "2019-04-01T10:00:00Z",
        "attempt_id": "0", "from_answer_id": "200", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "501", "type": "Task", "string": {"title": "Single"}}
      },
      {
        "activity_type": "result_validated", "at": null,
        "attempt_id": "0", "from_answer_id": "200", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "501", "type": "Task", "string": {"title": "Single"}}
      }
    ]
    """
