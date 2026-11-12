Feature: Activity log filtered by attempt on a task

  Background:
    Given the database has the following users:
      | login | group_id | default_language | profile                                      |
      | user  | 11       | en               | {"first_name": "John", "last_name": "Doe"}   |
      | owner | 21       | en               | {"first_name": "Jean", "last_name": "Owner"} |
      | jane  | 31       | en               | {"first_name": "Jane", "last_name": "Doe"}   |
    And the database has the following table "groups":
      | id | type  | name      |
      | 13 | Class | Our Class |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 13       | 21         | true              |
      | 11       | 21         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 13              | 11             | 2019-05-30 11:00:00            |
      | 13              | 31             | 2019-05-30 11:00:00            |
    And the groups ancestors are computed
    And the database has the following table "languages":
      | tag |
      | en  |
    And the database has the following table "items":
      | id  | type | no_score | default_language_tag | allows_multiple_attempts |
      | 500 | Task | false    | en                   | 1                        |
    And the database has the following table "items_strings":
      | item_id | language_tag | title |
      | 500     | en           | Multi |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 11       | 500     | content            | none                |
      | 21       | 500     | info               | answer              |
      | 31       | 500     | content            | none                |
    And the database has the following table "attempts":
      | participant_id | id | parent_attempt_id | root_item_id |
      | 11             | 0  | null              | null         |
      | 11             | 1  | 0                 | 500          |
      | 11             | 2  | 0                 | 500          |
      | 31             | 0  | null              | null         |
      | 31             | 1  | 0                 | 500          |
    And the database has the following table "results":
      | participant_id | attempt_id | item_id | started_at          | validated_at        |
      | 11             | 0          | 500     | 2019-05-28 10:00:00 | null                |
      | 11             | 1          | 500     | 2019-05-29 10:00:00 | 2019-05-29 14:00:00 |
      | 11             | 2          | 500     | 2019-05-30 10:00:00 | null                |
      | 31             | 0          | 500     | 2019-06-02 10:00:00 | null                |
      | 31             | 1          | 500     | 2019-06-03 10:00:00 | null                |
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | created_at          |
      | 100 | 11        | 11             | 0          | 500     | Submission | 2019-05-28 11:00:00 |
      | 101 | 11        | 11             | 1          | 500     | Submission | 2019-05-29 11:00:00 |
      | 102 | 11        | 11             | 1          | 500     | Saved      | 2019-05-29 12:00:00 |
      | 103 | 11        | 11             | 1          | 500     | Current    | 2019-05-29 13:00:00 |
      | 104 | 11        | 11             | 2          | 500     | Submission | 2019-05-30 11:00:00 |
      | 300 | 31        | 31             | 0          | 500     | Submission | 2019-06-02 11:00:00 |
      | 301 | 31        | 31             | 1          | 500     | Submission | 2019-06-03 11:00:00 |

  Scenario: Participant's own log filtered to one attempt
    Given I am the user with id "11"
    When I send a GET request to "/items/500/log?attempt_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated", "at": "2019-05-29T14:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "-1",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "current_answer", "answer_id": "103", "at": "2019-05-29T13:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "103",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "saved_answer", "answer_id": "102", "at": "2019-05-29T12:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "102",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "101", "at": "2019-05-29T11:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "101",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-29T10:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "101",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      }
    ]
    """

  Scenario: Filter keeps pagination and tie-breaking within the attempt
    Given I am the user with id "11"
    When I send a GET request to "/items/500/log?attempt_id=1&limit=2&from.activity_type=current_answer&from.participant_id=11&from.attempt_id=1&from.item_id=500&from.answer_id=103"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "saved_answer", "answer_id": "102", "at": "2019-05-29T12:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "102",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "101", "at": "2019-05-29T11:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "101",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      }
    ]
    """

  Scenario: Watching a user filters that user's attempt
    Given I am the user with id "21"
    When I send a GET request to "/items/500/log?watched_group_id=11&attempt_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated", "at": "2019-05-29T14:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "-1", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "current_answer", "answer_id": "103", "at": "2019-05-29T13:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "103", "can_watch_answer": true,
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
      }
    ]
    """
