Feature: Get activity log
  Background:
    Given the database has the following users:
      | login | temp_user | group_id | first_name  | last_name | default_language |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | fr               |
      | user  | 0         | 11       | John        | Doe       | en               |
      | jane  | 0         | 31       | Jane        | Doe       | en               |
      | paul  | 0         | 41       | Paul        | Smith     | en               |
    And the database has the following table 'groups':
      | id | type  | name       |
      | 13 | Class | Our Class  |
      | 20 | Other | Some Group |
      | 30 | Team  | Our Team   |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 11       | 31         | true              |
      | 13       | 21         | true              |
      | 31       | 31         | true              |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 13              | 11             | 2020-01-01 00:00:00            |
      | 13              | 41             | 2020-01-01 00:00:00            |
      | 20              | 21             | null                           |
      | 30              | 21             | null                           |
    And the groups ancestors are computed
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 11             |
      | 1  | 11             |
    And the database has the following table 'results':
      | attempt_id | item_id | participant_id | started_at          | validated_at        | latest_submission_at |
      | 0          | 200     | 30             | 2020-01-01 00:09:00 | 2020-01-01 00:10:00 | 2020-01-01 00:10:00  |
      | 1          | 200     | 31             | 2020-01-01 00:19:00 | 2020-01-01 00:20:00 | 2020-01-01 00:20:00  |
      | 1          | 200     | 41             | 2020-01-01 00:28:00 | 2020-01-01 00:29:00 | 2020-01-01 00:29:00  |
      | 0          | 202     | 11             | 2020-01-01 00:30:00 | 2020-01-01 00:57:00 | 2020-01-01 00:57:00  |
      | 1          | 202     | 11             | 2020-01-01 00:31:00 | 2020-01-01 00:58:00 | 2020-01-01 00:58:00  |
      | 0          | 201     | 11             | 2020-01-01 00:32:00 | null                | 2020-01-01 00:32:00  |
      | 1          | 200     | 11             | 2020-01-01 00:33:00 | 2020-01-01 00:59:00 | 2020-01-01 00:59:00  |
      | 1          | 201     | 11             | null                | 2020-01-01 00:40:00 | 2020-01-01 00:40:00  |
      | 0          | 200     | 11             | 2020-01-01 00:41:00 | 2020-01-01 00:44:00 | 2020-01-01 00:44:00  |
      | 0          | 203     | 11             | 2020-01-01 01:00:00 | 2020-01-01 01:00:00 | 2020-01-01 01:00:00  |
      | 1          | 203     | 11             | 2020-01-01 01:00:00 | 2020-01-01 01:00:00 | 2020-01-01 01:00:00  |
    And the database has the following table 'answers':
      | id | author_id | participant_id | attempt_id | item_id | type       | state   | created_at          |
      | 1  | 31        | 11             | 0          | 202     | Submission | State1  | 2020-01-01 00:34:00 |
      | 2  | 11        | 11             | 0          | 202     | Submission | State2  | 2020-01-01 00:35:00 |
      | 3  | 31        | 11             | 1          | 202     | Submission | State3  | 2020-01-01 00:36:00 |
      | 4  | 11        | 11             | 1          | 202     | Submission | State4  | 2020-01-01 00:37:00 |
      | 5  | 31        | 11             | 0          | 201     | Submission | State5  | 2020-01-01 00:38:00 |
      | 6  | 11        | 11             | 0          | 201     | Submission | State6  | 2020-01-01 00:39:00 |
      | 7  | 31        | 11             | 0          | 200     | Submission | State7  | 2020-01-01 00:42:00 |
      | 8  | 11        | 11             | 0          | 200     | Submission | State8  | 2020-01-01 00:43:00 |
      | 9  | 31        | 11             | 1          | 200     | Submission | State9  | 2020-01-01 00:45:00 |
      | 10 | 11        | 11             | 1          | 200     | Submission | State10 | 2020-01-01 00:46:00 |
      | 11 | 31        | 11             | 1          | 202     | Submission | State11 | 2020-01-01 00:47:00 |
      | 12 | 11        | 11             | 1          | 202     | Submission | State12 | 2020-01-01 00:48:00 |
      | 13 | 11        | 11             | 1          | 202     | Saved      | State13 | 2020-01-01 00:49:00 |
      | 14 | 11        | 11             | 1          | 202     | Current    | State14 | 2020-01-01 00:50:00 |
      | 15 | 11        | 11             | 1          | 201     | Saved      | State15 | 2020-01-01 00:51:00 |
      | 16 | 11        | 11             | 1          | 201     | Current    | State16 | 2020-01-01 00:52:00 |
      | 17 | 41        | 41             | 1          | 200     | Submission | State17 | 2020-01-01 00:53:00 |
      | 18 | 31        | 11             | 1          | 200     | Submission | State18 | 2020-01-01 00:54:00 |
      | 19 | 11        | 11             | 1          | 200     | Saved      | State19 | 2020-01-01 00:55:00 |
      | 20 | 11        | 11             | 1          | 200     | Current    | State20 | 2020-01-01 00:56:00 |
      | 21 | 11        | 11             | 0          | 203     | Submission | State21 | 2020-01-01 01:00:00 |
      | 22 | 11        | 11             | 1          | 203     | Submission | State22 | 2020-01-01 01:00:01 |
      | 23 | 11        | 11             | 1          | 203     | Submission | State23 | 2020-01-01 01:00:02 |
      | 24 | 11        | 11             | 1          | 203     | Saved      | State24 | 2020-01-01 01:00:03 |
      | 25 | 11        | 11             | 1          | 203     | Current    | State25 | 2020-01-01 01:00:04 |
      | 26 | 31        | 11             | 1          | 203     | Submission | State26 | 2020-01-01 01:00:05 |
      | 27 | 31        | 11             | 0          | 203     | Submission | State27 | 2020-01-01 01:00:06 |
      | 28 | 31        | 11             | 1          | 203     | Submission | State28 | 2020-01-01 01:00:07 |
      | 29 | 31        | 11             | 1          | 204     | Submission | State29 | 2020-01-01 01:00:08 |
    And the database has the following table 'gradings':
      | answer_id | graded_at           | score |
      | 5         | 2020-01-01 00:38:00 | 98    |
      | 6         | 2020-01-01 00:39:00 | 99    |
      | 15        | 2020-01-01 00:51:00 | 100   |
      | 16        | 2020-01-01 00:52:00 | 100   |
      | 50        | 2020-01-01 01:00:00 | 100   |
      | 51        | 2020-01-01 01:00:00 | 100   |
      | 52        | 2020-01-01 01:00:00 | 100   |
      | 53        | 2020-01-01 01:00:00 | 100   |
    And the database has the following table 'items':
      | id  | type    | no_score | default_language_tag |
      | 200 | Task    | false    | fr                   |
      | 201 | Chapter | false    | fr                   |
      | 202 | Chapter | false    | fr                   |
      | 203 | Chapter | false    | fr                   |
      | 204 | Task    | false    | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 20       | 200     | none               | result              |
      | 21       | 200     | info               | none                | # user 21 is in group 20, so he has can_watch=result on item 200
      | 21       | 201     | info               | result              |
      | 21       | 202     | info               | result              |
      | 21       | 203     | none               | result              |
      | 21       | 204     | content            | none                |
      | 30       | 200     | content            | answer              |
      | 31       | 200     | content            | answer              |
      | 31       | 201     | content            | answer              |
      | 31       | 202     | content            | answer              |
      | 31       | 203     | content            | none                |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 200              | 201           |
      | 200              | 203           |
      | 200              | 204           |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title      | image_url                  | subtitle     | description   | edu_comment    |
      | 200     | en           | Task 1     | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 200     | fr           | Tache 1    | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 201     | en           | Chapter 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 201     | fr           | Chapitre 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 202     | en           | Chapter 2  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 202     | fr           | Chapitre 2 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 203     | en           | Chapter 3  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 203     | fr           | Chapitre 3 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
    And the database has the following table 'languages':
      | tag |
      | fr  |

  Scenario Outline: User is a manager of the group and there are visible descendants of the item
      This spec also checks:
      1) activities ordering,
      2) filtering by users groups,
      3) that a user cannot see names of other users without approval
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/200/log?watched_group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:59:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "-1"
      },
      {
        "activity_type": "current_answer",
        "answer_id": "20",
        "at": "2020-01-01T00:56:00Z",
        "attempt_id": "1",
        "from_answer_id": "20",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "19",
        "at": "2020-01-01T00:55:00Z",
        "attempt_id": "1",
        "from_answer_id": "19",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:54:00Z",
        "answer_id": "18",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "18"
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:53:00Z",
        "answer_id": "17",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "17"
      },
      {
        "activity_type": "current_answer",
        "answer_id": "16",
        "at": "2020-01-01T00:52:00Z",
        "attempt_id": "1",
        "from_answer_id": "16",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 100,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "15",
        "at": "2020-01-01T00:51:00Z",
        "attempt_id": "1",
        "from_answer_id": "15",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 100,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:46:00Z",
        "answer_id": "10",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "10"
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:45:00Z",
        "answer_id": "9",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "9"
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:44:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "9"
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:43:00Z",
        "answer_id": "8",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "8"
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:42:00Z",
        "answer_id": "7",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "7"
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:41:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "7"
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:40:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "7"
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:39:00Z",
        "answer_id": "6",
        "score": 99,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "6"
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:38:00Z",
        "answer_id": "5",
        "score": 98,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "5"
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:33:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "can_watch_item_answer": false,
        "from_answer_id": "5"
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:32:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "5"
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:29:00Z",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "5"
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:28:00Z",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "5"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: User is a manager of the group and there are visible descendants of the item; request the first row
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/200/log?watched_group_id=13&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:59:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "-1"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: User is a manager of the group and there are visible descendants of the item; request two rows right after a row with activity_type="result_validated"
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=result_validated&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=-1&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "current_answer",
        "answer_id": "20",
        "at": "2020-01-01T00:56:00Z",
        "attempt_id": "1",
        "from_answer_id": "20",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "19",
        "at": "2020-01-01T00:55:00Z",
        "attempt_id": "1",
        "from_answer_id": "19",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: User is a manager of the group and there are visible descendants of the item; request two rows right after a row with activity_type="current_answer"
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=current_answer&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=20&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "saved_answer",
        "answer_id": "19",
        "at": "2020-01-01T00:55:00Z",
        "attempt_id": "1",
        "from_answer_id": "19",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:54:00Z",
        "answer_id": "18",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "18"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: User is a manager of the group and there are visible descendants of the item; request two rows right after a row with activity_type="saved_answer"
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=saved_answer&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=19&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:54:00Z",
        "answer_id": "18",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "18"
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:53:00Z",
        "answer_id": "17",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "17"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: User is a manager of the group and there are visible descendants of the item; request two rows right after a row with activity_type="submission"
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=submission&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=9&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:44:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "9"
      },
      {
        "activity_type": "submission",
        "at": "2020-01-01T00:43:00Z",
        "answer_id": "8",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "8"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: User is a manager of the group and there are visible descendants of the item; request a row right after a row with activity_type="result_started"
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=result_started&from.participant_id=11&from.attempt_id=0&from.item_id=200&from.answer_id=7&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:40:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "7"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: User is a manager of the group and there are visible descendants of the item; request the last rows
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=result_started&from.participant_id=11&from.attempt_id=0&from.item_id=201&from.answer_id=5"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:29:00Z",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "5"
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:28:00Z",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "5"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario: User can see their own name
    Given I am the user with id "31"
    When I send a GET request to "/items/200/log?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:20:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_item_answer": true,
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "from_answer_id": "-1"
      }
    ]
    """

  Scenario: A user can view activity of his team
    Given I am the user with id "21"
    When I send a GET request to "/items/200/log?as_team_id=30"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:10:00Z",
        "participant": {"id": "30", "name": "Our Team", "type": "Team"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "from_answer_id": "-1"
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:09:00Z",
        "participant": {"id": "30", "name": "Our Team", "type": "Team"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "from_answer_id": "-1"
      }
    ]
    """

  Scenario Outline: Get activity for all visible items, the user is a manager of the watched group
      This spec also checks:
      1) activities ordering,
      2) filtering by users groups,
      3) that a user cannot see names of other users without approval
    Given I am the user with id "21"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/log?watched_group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:59:00Z",
        "attempt_id": "1",
        "from_answer_id": "-1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:58:00Z",
        "attempt_id": "1",
        "from_answer_id": "-1",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:57:00Z",
        "attempt_id": "0",
        "from_answer_id": "-1",
        "item": {"id": "202", "string": { "title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "current_answer",
        "answer_id": "20",
        "at": "2020-01-01T00:56:00Z",
        "attempt_id": "1",
        "from_answer_id": "20",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "19",
        "at": "2020-01-01T00:55:00Z",
        "attempt_id": "1",
        "from_answer_id": "19",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "18",
        "at": "2020-01-01T00:54:00Z",
        "attempt_id": "1",
        "from_answer_id": "18",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "submission",
        "answer_id": "17",
        "at": "2020-01-01T00:53:00Z",
        "attempt_id": "1",
        "from_answer_id": "17",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "user": {"first_name": "Paul", "id": "41", "last_name": "Smith", "login": "paul"}
      },
      {
        "activity_type": "current_answer",
        "answer_id": "16",
        "at": "2020-01-01T00:52:00Z",
        "attempt_id": "1",
        "from_answer_id": "16",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 100,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "15",
        "at": "2020-01-01T00:51:00Z",
        "attempt_id": "1",
        "from_answer_id": "15",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 100,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "current_answer",
        "answer_id": "14",
        "at": "2020-01-01T00:50:00Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "13",
        "at": "2020-01-01T00:49:00Z",
        "attempt_id": "1",
        "from_answer_id": "13",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "12",
        "at": "2020-01-01T00:48:00Z",
        "attempt_id": "1",
        "from_answer_id": "12",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "11",
        "at": "2020-01-01T00:47:00Z",
        "attempt_id": "1",
        "from_answer_id": "11",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "submission",
        "answer_id": "10",
        "at": "2020-01-01T00:46:00Z",
        "attempt_id": "1",
        "from_answer_id": "10",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "9",
        "at": "2020-01-01T00:45:00Z",
        "attempt_id": "1",
        "from_answer_id": "9",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:44:00Z",
        "attempt_id": "0",
        "from_answer_id": "9",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "8",
        "at": "2020-01-01T00:43:00Z",
        "attempt_id": "0",
        "from_answer_id": "8",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "7",
        "at": "2020-01-01T00:42:00Z",
        "attempt_id": "0",
        "from_answer_id": "7",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:41:00Z",
        "attempt_id": "0",
        "from_answer_id": "7",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:40:00Z",
        "attempt_id": "1",
        "from_answer_id": "7",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "6",
        "at": "2020-01-01T00:39:00Z",
        "attempt_id": "0",
        "from_answer_id": "6",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 99,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "5",
        "at": "2020-01-01T00:38:00Z",
        "attempt_id": "0",
        "from_answer_id": "5",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 98,
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "submission",
        "answer_id": "4",
        "at": "2020-01-01T00:37:00Z",
        "attempt_id": "1",
        "from_answer_id": "4",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "3",
        "at": "2020-01-01T00:36:00Z",
        "attempt_id": "1",
        "from_answer_id": "3",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "submission",
        "answer_id": "2",
        "at": "2020-01-01T00:35:00Z",
        "attempt_id": "0",
        "from_answer_id": "2",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "1",
        "at": "2020-01-01T00:34:00Z",
        "attempt_id": "0",
        "from_answer_id": "1",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:33:00Z",
        "attempt_id": "1",
        "from_answer_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:32:00Z",
        "attempt_id": "0",
        "from_answer_id": "1",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:31:00Z",
        "attempt_id": "1",
        "from_answer_id": "1",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:30:00Z",
        "attempt_id": "0",
        "from_answer_id": "1",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "can_watch_item_answer": false,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:29:00Z",
        "attempt_id": "1",
        "from_answer_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "user": {"first_name": "Paul", "id": "41", "last_name": "Smith", "login": "paul"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:28:00Z",
        "attempt_id": "1",
        "from_answer_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "can_watch_item_answer": false,
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "user": {"first_name": "Paul", "id": "41", "last_name": "Smith", "login": "paul"}
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: Get activity of the current user for all visible items
    Given I am the user with id "31"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:20:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_item_answer": true,
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "from_answer_id": "-1"
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:19:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_item_answer": true,
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "from_answer_id": "-1"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: Get activity of the current user for all visible items (only the first row)
    Given I am the user with id "31"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/log?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:20:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_item_answer": true,
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "from_answer_id": "-1"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |

  Scenario Outline: Get activity of the current user for all visible items (start from the second row)
    Given I am the user with id "31"
    And the context variable "forceStraightJoinInItemActivityLog" is "<forceStraightJoinInItemActivityLog>"
    When I send a GET request to "/items/log?from.activity_type=result_validated&from.participant_id=31&from.attempt_id=1&from.item_id=200&from.answer_id=-1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:19:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "can_watch_item_answer": true,
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "from_answer_id": "-1"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |
