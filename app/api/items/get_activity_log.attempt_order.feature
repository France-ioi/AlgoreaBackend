Feature: Activity log attempt order

  Background:
    Given the database has the following users:
      | login | group_id | default_language | profile                                    |
      | user  | 11       | en               | {"first_name": "John", "last_name": "Doe"} |
      | owner | 21       | en               | {"first_name": "Jean", "last_name": "Owner"} |
      | jane  | 31       | en               | {"first_name": "Jane", "last_name": "Doe"} |
    And the database has the following table "groups":
      | id | type  | name      |
      | 13 | Class | Our Class |
      | 30 | Team  | Our Team  |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 13       | 21         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 13              | 11             | 2019-05-30 11:00:00            |
      | 13              | 31             | 2019-05-30 11:00:00            |
      | 30              | 21             | null                           |
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
      | 11       | 500     | content            | none                |
      | 11       | 501     | content            | none                |
      | 21       | 500     | info               | answer              |
      | 21       | 501     | info               | answer              |
      | 30       | 500     | content            | none                |
      | 31       | 500     | content            | none                |
    And the database has the following table "attempts":
      | participant_id | id | parent_attempt_id | root_item_id |
      | 11             | 0  | null              | null         |
      | 11             | 1  | 0                 | 500          |
      | 11             | 2  | 0                 | 500          |
      | 11             | 3  | 0                 | 500          |
      | 11             | 4  | 0                 | 500          |
      | 11             | 10 | 0                 | 500          |
      | 11             | 11 | 10                | 500          |
      | 31             | 0  | null              | null         |
      | 31             | 1  | 0                 | 500          |
      | 30             | 0  | null              | null         |
      | 30             | 1  | 0                 | 500          |
    And the database has the following table "results":
      | participant_id | attempt_id | item_id | started_at          | validated_at        |
      | 11             | 0          | 500     | 2019-05-28 10:00:00 | null                |
      | 11             | 1          | 500     | 2019-05-29 10:00:00 | 2019-05-29 14:00:00 |
      | 11             | 2          | 500     | 2019-05-30 10:00:00 | null                |
      | 11             | 3          | 500     | null                | null                |
      | 11             | 4          | 500     | 2019-05-30 10:00:00 | null                |
      | 11             | 10         | 500     | 2019-05-31 10:00:00 | null                |
      | 11             | 11         | 500     | 2019-06-01 10:00:00 | null                |
      | 11             | 0          | 501     | 2019-04-01 10:00:00 | null                |
      | 31             | 0          | 500     | 2019-06-02 10:00:00 | null                |
      | 31             | 1          | 500     | 2019-06-03 10:00:00 | null                |
      | 30             | 0          | 500     | 2019-03-01 10:00:00 | null                |
      | 30             | 1          | 500     | 2019-03-02 10:00:00 | null                |
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | created_at          |
      | 100 | 11        | 11             | 0          | 500     | Submission | 2019-05-28 11:00:00 |
      | 101 | 11        | 11             | 1          | 500     | Submission | 2019-05-29 11:00:00 |
      | 102 | 11        | 11             | 1          | 500     | Saved      | 2019-05-29 12:00:00 |
      | 103 | 11        | 11             | 1          | 500     | Current    | 2019-05-29 13:00:00 |
      | 104 | 11        | 11             | 2          | 500     | Submission | 2019-05-30 11:00:00 |
      | 105 | 11        | 11             | 4          | 500     | Submission | 2019-05-30 11:00:00 |
      | 106 | 11        | 11             | 10         | 500     | Submission | 2019-05-31 11:00:00 |
      | 107 | 11        | 11             | 11         | 500     | Submission | 2019-06-01 11:00:00 |
      | 200 | 11        | 11             | 0          | 501     | Submission | 2019-04-01 11:00:00 |
      | 300 | 31        | 31             | 0          | 500     | Submission | 2019-06-02 11:00:00 |
      | 301 | 31        | 31             | 1          | 500     | Submission | 2019-06-03 11:00:00 |
      | 400 | 21        | 30             | 0          | 500     | Submission | 2019-03-01 11:00:00 |
      | 401 | 21        | 30             | 1          | 500     | Submission | 2019-03-02 11:00:00 |

  Scenario: Item log ranks started siblings, ignores an unstarted one, and breaks ties by attempt id
    Given I am the user with id "11"
    When I send a GET request to "/items/500/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission", "answer_id": "107", "at": "2019-06-01T11:00:00Z",
        "attempt_id": "11", "attempt_order": 2, "from_answer_id": "107",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-06-01T10:00:00Z",
        "attempt_id": "11", "attempt_order": 2, "from_answer_id": "107",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "106", "at": "2019-05-31T11:00:00Z",
        "attempt_id": "10", "attempt_order": 5, "from_answer_id": "106",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-31T10:00:00Z",
        "attempt_id": "10", "attempt_order": 5, "from_answer_id": "106",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "105", "at": "2019-05-30T11:00:00Z",
        "attempt_id": "4", "attempt_order": 4, "from_answer_id": "105",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "104", "at": "2019-05-30T11:00:00Z",
        "attempt_id": "2", "attempt_order": 3, "from_answer_id": "104",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-30T10:00:00Z",
        "attempt_id": "4", "attempt_order": 4, "from_answer_id": "104",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-30T10:00:00Z",
        "attempt_id": "2", "attempt_order": 3, "from_answer_id": "104",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_validated", "at": "2019-05-29T14:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "104",
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
      },
      {
        "activity_type": "submission", "answer_id": "100", "at": "2019-05-28T11:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "100",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-28T10:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "100",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      }
    ]
    """

  Scenario: A single-attempt item omits attempt_order
    Given I am the user with id "11"
    When I send a GET request to "/items/501/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission", "answer_id": "200", "at": "2019-04-01T11:00:00Z",
        "attempt_id": "0", "from_answer_id": "200",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "501", "type": "Task", "string": {"title": "Single"}}
      },
      {
        "activity_type": "result_started", "at": "2019-04-01T10:00:00Z",
        "attempt_id": "0", "from_answer_id": "200",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "501", "type": "Task", "string": {"title": "Single"}}
      }
    ]
    """

  Scenario: All-items log uses the same attempt orders and still omits them on a single-attempt item
    Given I am the user with id "11"
    When I send a GET request to "/items/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission", "answer_id": "107", "at": "2019-06-01T11:00:00Z",
        "attempt_id": "11", "attempt_order": 2, "from_answer_id": "107",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-06-01T10:00:00Z",
        "attempt_id": "11", "attempt_order": 2, "from_answer_id": "107",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "106", "at": "2019-05-31T11:00:00Z",
        "attempt_id": "10", "attempt_order": 5, "from_answer_id": "106",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-31T10:00:00Z",
        "attempt_id": "10", "attempt_order": 5, "from_answer_id": "106",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "105", "at": "2019-05-30T11:00:00Z",
        "attempt_id": "4", "attempt_order": 4, "from_answer_id": "105",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "104", "at": "2019-05-30T11:00:00Z",
        "attempt_id": "2", "attempt_order": 3, "from_answer_id": "104",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-30T10:00:00Z",
        "attempt_id": "4", "attempt_order": 4, "from_answer_id": "104",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-30T10:00:00Z",
        "attempt_id": "2", "attempt_order": 3, "from_answer_id": "104",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_validated", "at": "2019-05-29T14:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "104",
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
      },
      {
        "activity_type": "submission", "answer_id": "100", "at": "2019-05-28T11:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "100",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-28T10:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "100",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "200", "at": "2019-04-01T11:00:00Z",
        "attempt_id": "0", "from_answer_id": "200",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "501", "type": "Task", "string": {"title": "Single"}}
      },
      {
        "activity_type": "result_started", "at": "2019-04-01T10:00:00Z",
        "attempt_id": "0", "from_answer_id": "200",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "501", "type": "Task", "string": {"title": "Single"}}
      }
    ]
    """

  Scenario: Watched-group rows are numbered per participant
    Given I am the user with id "21"
    When I send a GET request to "/items/500/log?watched_group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission", "answer_id": "301", "at": "2019-06-03T11:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "301", "can_watch_answer": true,
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-06-03T10:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "301", "can_watch_answer": true,
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "300", "at": "2019-06-02T11:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "300", "can_watch_answer": true,
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-06-02T10:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "300", "can_watch_answer": true,
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "107", "at": "2019-06-01T11:00:00Z",
        "attempt_id": "11", "attempt_order": 2, "from_answer_id": "107", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-06-01T10:00:00Z",
        "attempt_id": "11", "attempt_order": 2, "from_answer_id": "107", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "106", "at": "2019-05-31T11:00:00Z",
        "attempt_id": "10", "attempt_order": 5, "from_answer_id": "106", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-05-31T10:00:00Z",
        "attempt_id": "10", "attempt_order": 5, "from_answer_id": "106", "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
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
      }
    ]
    """

  Scenario: Team log numbers the team's attempts
    Given I am the user with id "21"
    When I send a GET request to "/items/500/log?as_team_id=30"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission", "answer_id": "401", "at": "2019-03-02T11:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "401",
        "participant": {"id": "30", "name": "Our Team", "type": "Team"},
        "user": {"id": "21", "first_name": "Jean", "last_name": "Owner", "login": "owner"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-03-02T10:00:00Z",
        "attempt_id": "1", "attempt_order": 2, "from_answer_id": "401",
        "participant": {"id": "30", "name": "Our Team", "type": "Team"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "submission", "answer_id": "400", "at": "2019-03-01T11:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "400",
        "participant": {"id": "30", "name": "Our Team", "type": "Team"},
        "user": {"id": "21", "first_name": "Jean", "last_name": "Owner", "login": "owner"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      },
      {
        "activity_type": "result_started", "at": "2019-03-01T10:00:00Z",
        "attempt_id": "0", "attempt_order": 1, "from_answer_id": "400",
        "participant": {"id": "30", "name": "Our Team", "type": "Team"},
        "item": {"id": "500", "type": "Task", "string": {"title": "Multi"}}
      }
    ]
    """
