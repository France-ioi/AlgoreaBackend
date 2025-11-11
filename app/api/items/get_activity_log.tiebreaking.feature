Feature: Get activity log
  Background:
    Given the database has the following users:
      | login | group_id | default_language | profile                                                |
      | owner | 21       | fr               | {"first_name": "Jean-Michel", "last_name": "Blanquer"} |
      | user  | 11       | en               | {"first_name": "John", "last_name": "Doe"}             |
    And the database has the following table "groups":
      | id | type  | name       |
      | 13 | Class | Our Class  |
      | 20 | Other | Some Group |
      | 30 | Team  | Our Team   |
      | 40 | Club  | Our Club   |
      | 50 | Club  | Team Club  |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 13       | 21         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 13              | 11             | 2019-05-30 11:00:00            |
      | 20              | 21             | null                           |
      | 30              | 21             | null                           |
      | 40              | 21             | null                           |
      | 50              | 30             | null                           |
    And the groups ancestors are computed
    And the database has the following table "attempts":
      | id | participant_id |
      | 0  | 11             |
      | 1  | 11             |
    And the database has the following table "results":
      | attempt_id | item_id | participant_id | started_at          | validated_at        | latest_submission_at |
      | 1          | 200     | 11             | 2017-05-29 06:38:38 | 2017-05-29 06:38:38 | 2020-05-29 06:38:38  |
    And the database has the following table "answers":
      | id | author_id | participant_id | attempt_id | item_id | type       | state   | created_at          |
      | 11 | 11        | 11             | 1          | 200     | Submission | State11 | 2017-05-29 06:38:38 |
      | 12 | 11        | 11             | 1          | 200     | Submission | State12 | 2017-05-29 06:38:38 |
      | 14 | 11        | 11             | 1          | 200     | Saved      | State14 | 2017-05-29 06:38:38 |
      | 15 | 11        | 11             | 1          | 200     | Current    | State15 | 2017-05-29 06:38:38 |
    And the database has the following table "items":
      | id  | type    | no_score | default_language_tag |
      | 200 | Task    | false    | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 20       | 200     | none               | result              |
      | 21       | 200     | info               | none                | # user 21 is in group 20, so he has can_watch=result on item 200
      | 30       | 200     | content            | answer              |
    And the database has the following table "items_strings":
      | item_id | language_tag | title      | image_url                  | subtitle     | description   | edu_comment    |
      | 200     | en           | Task 1     | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 200     | fr           | Tache 1    | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
    And the database has the following table "languages":
      | tag |
      | fr  |

  Scenario: Activity log with items having equal values in all columns except activity_type and answer_id
    Given I am the user with id "21"
    When I send a GET request to "/items/200/log?watched_group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "current_answer",
        "answer_id": "15",
        "at": "2017-05-29T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "15",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "14",
        "at": "2017-05-29T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "14"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "11",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "11"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "12",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "12"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "can_watch_answer": false,
        "from_answer_id": "12"
      }
    ]
    """

  Scenario: Activity log with items having equal values in all columns except activity_type and answer_id; second row
    Given I am the user with id "21"
    When I send a GET request to "/items/200/log?watched_group_id=13&limit=1&from.activity_type=current_answer&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=15"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "saved_answer",
        "answer_id": "14",
        "at": "2017-05-29T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      }
    ]
    """

  Scenario: Activity log with items having equal values in all columns except activity_type and answer_id; third row
    Given I am the user with id "21"
    When I send a GET request to "/items/200/log?watched_group_id=13&limit=1&from.activity_type=saved_answer&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "14"
      }
    ]
    """

  Scenario: Activity log with items having equal values in all columns except activity_type and answer_id; fourth row
    Given I am the user with id "21"
    When I send a GET request to "/items/200/log?watched_group_id=13&limit=1&from.activity_type=result_validated&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "11",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "11"
      }
    ]
    """

  Scenario: Activity log with items having equal values in all columns except activity_type and answer_id; fifth row
    Given I am the user with id "21"
    When I send a GET request to "/items/200/log?watched_group_id=13&limit=1&from.activity_type=submission&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "12",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "12"
      }
    ]
    """

  Scenario: Activity log with items having equal values in all columns except activity_type and answer_id; sixth row
    Given I am the user with id "21"
    When I send a GET request to "/items/200/log?watched_group_id=13&limit=1&from.activity_type=submission&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=12"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "can_watch_answer": false,
        "from_answer_id": "12"
      }
    ]
    """
