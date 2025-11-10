Feature: Get activity log for a thread
  Background:
    Given the database has the following users:
      | login | group_id | default_language | profile                                                |
      | owner | 21       | fr               | {"first_name": "Jean-Michel", "last_name": "Blanquer"} |
      | user  | 11       | en               | {"first_name": "John", "last_name": "Doe"}             |
      | jane  | 31       | en               | {"first_name": "Jane", "last_name": "Doe"}             |
      | paul  | 41       | en               | {"first_name": "Paul", "last_name": "Smith"}           |
    And the database has the following table "groups":
      | id | type  | name       |
      | 13 | Class | Our Class  |
      | 20 | Other | Some Group |
      | 30 | Team  | Our Team   |
      | 40 | Club  | Our Club   |
      | 50 | Club  | Team Club  |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 11       | 31         | true              |
      | 13       | 21         | true              |
      | 31       | 31         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 13              | 11             | 2019-05-30 11:00:00            |
      | 13              | 41             | 2019-05-30 11:00:00            |
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
      | 0          | 200     | 11             | 2017-05-29 06:38:38 | 2017-05-29 06:38:38 | 2020-05-29 06:38:38  |
      | 0          | 200     | 30             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 0          | 201     | 11             | 2017-05-29 06:38:00 | null                | 2020-05-29 06:38:38  |
      | 0          | 201     | 30             | 2017-05-29 06:37:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 0          | 202     | 11             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 0          | 203     | 11             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 0          | 204     | 30             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 1          | 200     | 11             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 1          | 200     | 31             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 1          | 200     | 41             | 2016-05-29 06:38:00 | 2016-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 1          | 201     | 11             | null                | 2017-05-29 06:38:38 | 2020-05-29 06:38:38  |
      | 1          | 202     | 11             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 1          | 203     | 11             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
    And the database has the following table "answers":
      | id | author_id | participant_id | attempt_id | item_id | type       | state   | created_at          |
      | 1  | 11        | 11             | 0          | 201     | Submission | State1  | 2017-05-29 06:38:38 |
      | 4  | 11        | 11             | 1          | 201     | Saved      | State4  | 2017-05-30 06:38:38 |
      | 5  | 11        | 11             | 1          | 201     | Current    | State5  | 2017-05-30 06:38:38 |
      | 7  | 31        | 11             | 0          | 201     | Submission | State7  | 2017-05-29 06:38:38 |
      | 11 | 11        | 11             | 0          | 200     | Submission | State11 | 2017-05-29 06:38:38 |
      | 12 | 11        | 11             | 1          | 200     | Submission | State12 | 2017-05-29 06:38:38 |
      | 13 | 41        | 41             | 1          | 200     | Submission | State13 | 2017-05-30 06:38:38 |
      | 14 | 11        | 11             | 1          | 200     | Saved      | State14 | 2017-05-30 06:38:38 |
      | 15 | 11        | 11             | 1          | 200     | Current    | State15 | 2017-05-30 06:38:38 |
      | 16 | 31        | 11             | 1          | 200     | Submission | State16 | 2017-05-29 06:38:38 |
      | 17 | 31        | 11             | 0          | 200     | Submission | State17 | 2017-05-29 06:38:38 |
      | 18 | 31        | 11             | 1          | 200     | Submission | State18 | 2017-05-30 06:38:38 |
      | 21 | 11        | 11             | 0          | 202     | Submission | State21 | 2017-05-29 06:38:38 |
      | 22 | 11        | 11             | 1          | 202     | Submission | State22 | 2017-05-29 06:38:38 |
      | 23 | 11        | 11             | 1          | 202     | Submission | State23 | 2017-05-30 06:38:38 |
      | 24 | 11        | 11             | 1          | 202     | Saved      | State24 | 2017-05-30 06:38:38 |
      | 25 | 11        | 11             | 1          | 202     | Current    | State25 | 2017-05-30 06:38:38 |
      | 26 | 31        | 11             | 1          | 202     | Submission | State26 | 2017-05-29 06:38:38 |
      | 27 | 31        | 11             | 0          | 202     | Submission | State27 | 2017-05-29 06:38:38 |
      | 28 | 31        | 11             | 1          | 202     | Submission | State28 | 2017-05-30 06:38:38 |
      | 31 | 11        | 11             | 0          | 203     | Submission | State31 | 2017-05-29 06:38:38 |
      | 32 | 11        | 11             | 1          | 203     | Submission | State32 | 2017-05-29 06:38:38 |
      | 33 | 11        | 11             | 1          | 203     | Submission | State33 | 2017-05-30 06:38:38 |
      | 34 | 11        | 11             | 1          | 203     | Saved      | State34 | 2017-05-30 06:38:38 |
      | 35 | 11        | 11             | 1          | 203     | Current    | State35 | 2017-05-30 06:38:38 |
      | 36 | 31        | 11             | 1          | 203     | Submission | State36 | 2017-05-29 06:38:38 |
      | 37 | 31        | 11             | 0          | 203     | Submission | State37 | 2017-05-29 06:38:38 |
      | 38 | 31        | 11             | 1          | 203     | Submission | State38 | 2017-05-30 06:38:38 |
      | 39 | 31        | 11             | 1          | 204     | Submission | State39 | 2017-05-30 06:38:38 |
    And the database has the following table "gradings":
      | answer_id | graded_at           | score |
      | 2         | 2017-05-29 06:38:38 | 100   |
      | 1         | 2017-05-29 06:38:38 | 99    |
      | 3         | 2017-05-30 06:38:38 | 100   |
      | 4         | 2017-05-30 06:38:38 | 100   |
      | 5         | 2017-05-30 06:38:38 | 100   |
      | 6         | 2017-05-29 06:38:38 | 100   |
      | 7         | 2017-05-29 06:38:38 | 98    |
      | 8         | 2017-05-30 06:38:38 | 100   |
    And the database has the following table "items":
      | id  | type    | no_score | default_language_tag |
      | 200 | Task    | false    | fr                   |
      | 201 | Chapter | false    | fr                   |
      | 202 | Chapter | false    | fr                   |
      | 203 | Chapter | false    | fr                   |
      | 204 | Task    | false    | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 20       | 200     | none               | result              |
      | 21       | 200     | info               | none                | # user 21 is in group 20, so he has can_watch=result on item 200
      | 21       | 201     | info               | result              |
      | 21       | 202     | info               | result              |
      | 21       | 203     | none               | result              |
      | 21       | 204     | content            | none                |
      | 30       | 200     | content            | answer              |
      | 30       | 201     | content            | result              |
      | 31       | 200     | content            | answer              |
      | 31       | 201     | content            | answer              |
      | 31       | 202     | content            | answer              |
      | 31       | 203     | content            | none                |
      | 40       | 201     | content            | answer              |
      | 50       | 204     | content            | answer              |
    And the database has the following table "items_ancestors":
      | ancestor_item_id | child_item_id |
      | 200              | 201           |
      | 200              | 203           |
      | 200              | 204           |
    And the database has the following table "items_strings":
      | item_id | language_tag | title      | image_url                  | subtitle     | description   | edu_comment    |
      | 200     | en           | Task 1     | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 200     | fr           | Tache 1    | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 201     | en           | Chapter 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 201     | fr           | Chapitre 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 202     | en           | Chapter 2  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 202     | fr           | Chapitre 2 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 203     | en           | Chapter 3  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 203     | fr           | Chapitre 3 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
    And the database has the following table "languages":
      | tag |
      | fr  |

  Scenario: Get a full activity log for a thread
    Given I am the user with id "11"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2017-05-30T12:00:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "-1"
      },
      {
        "activity_type": "current_answer",
        "answer_id": "15",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "15",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "14",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "at": "2017-05-30T06:38:38Z",
        "answer_id": "18",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "18"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "12",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "12"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "16",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "16"
      },
      {
        "activity_type": "result_validated",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "16"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "11",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "11"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "17",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "17"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "17"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "can_watch_answer": true,
        "from_answer_id": "17"
      }
    ]
    """

  Scenario: Get a full activity log for a thread; request the first row
    Given I am the user with id "11"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2017-05-30T12:00:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "-1"
      }
    ]
    """

  Scenario: Get a full activity log for a thread; request two rows right after a row with activity_type="result_validated"
    Given I am the user with id "11"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log?from.activity_type=result_validated&from.attempt_id=1&from.answer_id=-1&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "current_answer",
        "answer_id": "15",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "15",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "14",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      }
    ]
    """

  Scenario: Get a full activity log for a thread; request two rows right after a row with activity_type="current_answer"
    Given I am the user with id "11"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log?from.activity_type=current_answer&from.attempt_id=1&from.answer_id=15&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "saved_answer",
        "answer_id": "14",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "at": "2017-05-30T06:38:38Z",
        "answer_id": "18",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "18"
      }
    ]
    """

  Scenario: Get a full activity log for a thread; request two rows right after a row with activity_type="saved_answer"
    Given I am the user with id "11"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log?from.activity_type=saved_answer&from.attempt_id=1&from.answer_id=14&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission",
        "at": "2017-05-30T06:38:38Z",
        "answer_id": "18",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "18"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "12",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "12"
      }
    ]
    """

  Scenario: Get a full activity log for a thread; request two rows right after a row with activity_type="submission"
    Given I am the user with id "11"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log?from.activity_type=submission&from.attempt_id=1&from.answer_id=16&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "16"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "11",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "11"
      }
    ]
    """

  Scenario: Get a full activity log for a thread; request a row right after a row with activity_type="result_started"
    Given I am the user with id "11"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log?from.activity_type=result_started&from.attempt_id=0&from.answer_id=17&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "can_watch_answer": true,
        "from_answer_id": "17"
      }
    ]
    """

  Scenario: Get a full activity log for a thread; request the last rows
    Given I am the user with id "11"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log?from.activity_type=submission&from.attempt_id=0&from.answer_id=17"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_answer": true,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "17"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "can_watch_answer": true,
        "from_answer_id": "17"
      }
    ]
    """
