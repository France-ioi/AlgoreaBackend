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
      | 13              | 11             | 2019-05-30 11:00:00            |
      | 13              | 41             | 2019-05-30 11:00:00            |
      | 20              | 21             | null                           |
      | 30              | 21             | null                           |
    And the groups ancestors are computed
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 11             |
      | 1  | 11             |
    And the database has the following table 'results':
      | attempt_id | item_id | participant_id | started_at          | validated_at        | latest_submission_at |
      | 0          | 200     | 11             | 2020-01-01 00:41:00 | 2020-01-01 00:44:00 | 2020-05-29 06:38:38  |
      | 0          | 200     | 30             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 0          | 201     | 11             | 2020-01-01 00:32:00 | null                | 2020-05-29 06:38:38  |
      | 0          | 202     | 11             | 2020-01-01 00:30:00 | 2020-01-01 00:57:00 | 2020-05-29 06:38:38  |
      | 0          | 203     | 11             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 1          | 200     | 11             | 2020-01-01 00:33:00 | 2020-01-01 00:59:00 | 2020-05-29 06:38:38  |
      | 1          | 200     | 31             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
      | 1          | 200     | 41             | 2020-01-01 00:28:00 | 2020-01-01 00:29:00 | 2020-05-29 06:38:38  |
      | 1          | 201     | 11             | null                | 2020-01-01 00:40:00 | 2020-05-29 06:38:38  |
      | 1          | 202     | 11             | 2020-01-01 00:31:00 | 2020-01-01 00:58:00 | 2020-05-29 06:38:38  |
      | 1          | 203     | 11             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
    And the database has the following table 'answers':
      | id | author_id | participant_id | attempt_id | item_id | type       | state   | created_at          |
      | 1  | 11        | 11             | 0          | 201     | Submission | State1  | 2020-01-01 00:39:00 |
      | 4  | 11        | 11             | 1          | 201     | Saved      | State4  | 2020-01-01 00:51:00 |
      | 5  | 11        | 11             | 1          | 201     | Current    | State5  | 2020-01-01 00:52:00 |
      | 7  | 31        | 11             | 0          | 201     | Submission | State7  | 2020-01-01 00:38:00 |
      | 11 | 11        | 11             | 0          | 200     | Submission | State11 | 2020-01-01 00:43:00 |
      | 12 | 11        | 11             | 1          | 200     | Submission | State12 | 2020-01-01 00:46:00 |
      | 13 | 41        | 41             | 1          | 200     | Submission | State13 | 2020-01-01 00:53:00 |
      | 14 | 11        | 11             | 1          | 200     | Saved      | State14 | 2020-01-01 00:55:00 |
      | 15 | 11        | 11             | 1          | 200     | Current    | State15 | 2020-01-01 00:56:00 |
      | 16 | 31        | 11             | 1          | 200     | Submission | State16 | 2020-01-01 00:45:00 |
      | 17 | 31        | 11             | 0          | 200     | Submission | State17 | 2020-01-01 00:42:00 |
      | 18 | 31        | 11             | 1          | 200     | Submission | State18 | 2020-01-01 00:54:00 |
      | 21 | 11        | 11             | 0          | 202     | Submission | State21 | 2020-01-01 00:35:00 |
      | 22 | 11        | 11             | 1          | 202     | Submission | State22 | 2020-01-01 00:37:00 |
      | 23 | 11        | 11             | 1          | 202     | Submission | State23 | 2020-01-01 00:48:00 |
      | 24 | 11        | 11             | 1          | 202     | Saved      | State24 | 2020-01-01 00:49:00 |
      | 25 | 11        | 11             | 1          | 202     | Current    | State25 | 2020-01-01 00:50:00 |
      | 26 | 31        | 11             | 1          | 202     | Submission | State26 | 2020-01-01 00:36:00 |
      | 27 | 31        | 11             | 0          | 202     | Submission | State27 | 2020-01-01 00:34:00 |
      | 28 | 31        | 11             | 1          | 202     | Submission | State28 | 2020-01-01 00:47:00 |
      | 31 | 11        | 11             | 0          | 203     | Submission | State31 | 2017-05-29 06:38:38 |
      | 32 | 11        | 11             | 1          | 203     | Submission | State32 | 2017-05-29 06:38:38 |
      | 33 | 11        | 11             | 1          | 203     | Submission | State33 | 2017-05-30 06:38:38 |
      | 34 | 11        | 11             | 1          | 203     | Saved      | State34 | 2017-05-30 06:38:38 |
      | 35 | 11        | 11             | 1          | 203     | Current    | State35 | 2017-05-30 06:38:38 |
      | 36 | 31        | 11             | 1          | 203     | Submission | State36 | 2017-05-29 06:38:38 |
      | 37 | 31        | 11             | 0          | 203     | Submission | State37 | 2017-05-29 06:38:38 |
      | 38 | 31        | 11             | 1          | 203     | Submission | State38 | 2017-05-30 06:38:38 |
      | 39 | 31        | 11             | 1          | 204     | Submission | State39 | 2017-05-30 06:38:38 |
    And the database has the following table 'gradings':
      | answer_id | graded_at           | score |
      | 2         | 2017-05-29 06:38:38 | 100   |
      | 1         | 2017-05-29 06:38:38 | 99    |
      | 3         | 2017-05-30 06:38:38 | 100   |
      | 4         | 2017-05-30 06:38:38 | 100   |
      | 5         | 2017-05-30 06:38:38 | 100   |
      | 6         | 2017-05-29 06:38:38 | 100   |
      | 7         | 2017-05-29 06:38:38 | 98    |
      | 8         | 2017-05-30 06:38:38 | 100   |
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
      | 21       | 200     | info               | none                |
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
        "at": "2017-05-30T12:00:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "-1"
      },
      {
        "activity_type": "current_answer",
        "answer_id": "15",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "15",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "14",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "at": "2017-05-30T06:38:38Z",
        "answer_id": "18",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "18"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-30T06:38:38Z",
        "answer_id": "13",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "13"
      },
      {
        "activity_type": "current_answer",
        "answer_id": "5",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "5",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 100,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "4",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "4",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 100,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "12",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "12"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "16",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "16"
      },
      {
        "activity_type": "result_validated",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "16"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "11",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "11"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "17",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "17"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "17"
      },
      {
        "activity_type": "result_validated",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "17"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "1",
        "score": 99,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "1"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "7",
        "score": 98,
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "7"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "7"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "7"
      },
      {
        "activity_type": "result_validated",
        "at": "2016-05-30T12:00:00Z",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "7"
      },
      {
        "activity_type": "result_started",
        "at": "2016-05-29T06:38:00Z",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "7"
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
        "at": "2017-05-30T12:00:00Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
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
        "answer_id": "15",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "15",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "14",
        "at": "2017-05-30T06:38:38Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
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
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=current_answer&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=15&limit=2"
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
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "at": "2017-05-30T06:38:38Z",
        "answer_id": "18",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
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
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=saved_answer&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=14&limit=2"
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
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "31", "login": "jane"},
        "from_answer_id": "18"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-30T06:38:38Z",
        "answer_id": "13",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "13"
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
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=submission&from.participant_id=11&from.attempt_id=1&from.item_id=200&from.answer_id=16&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "16"
      },
      {
        "activity_type": "submission",
        "at": "2017-05-29T06:38:38Z",
        "answer_id": "11",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "11"
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
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=result_started&from.participant_id=11&from.attempt_id=0&from.item_id=200&from.answer_id=17&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2017-05-29T06:38:38Z",
        "participant": {"id": "11", "name": "user", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "user": {"id": "11", "first_name": "John", "last_name": "Doe", "login": "user"},
        "from_answer_id": "17"
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
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=result_started&from.participant_id=11&from.attempt_id=0&from.item_id=201&from.answer_id=7"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2016-05-30T12:00:00Z",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "7"
      },
      {
        "activity_type": "result_started",
        "at": "2016-05-29T06:38:00Z",
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "user": {"id": "41", "first_name": "Paul", "last_name": "Smith", "login": "paul"},
        "from_answer_id": "7"
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
        "at": "2017-05-30T12:00:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
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
        "at": "2017-05-30T12:00:00Z",
        "participant": {"id": "30", "name": "Our Team", "type": "Team"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "from_answer_id": "-1"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "participant": {"id": "30", "name": "Our Team", "type": "Team"},
        "attempt_id": "0",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
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
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:58:00Z",
        "attempt_id": "1",
        "from_answer_id": "-1",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:57:00Z",
        "attempt_id": "0",
        "from_answer_id": "-1",
        "item": {"id": "202", "string": { "title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "current_answer",
        "answer_id": "15",
        "at": "2020-01-01T00:56:00Z",
        "attempt_id": "1",
        "from_answer_id": "15",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "14",
        "at": "2020-01-01T00:55:00Z",
        "attempt_id": "1",
        "from_answer_id": "14",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
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
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "submission",
        "answer_id": "13",
        "at": "2020-01-01T00:53:00Z",
        "attempt_id": "1",
        "from_answer_id": "13",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "user": {"first_name": "Paul", "id": "41", "last_name": "Smith", "login": "paul"}
      },
      {
        "activity_type": "current_answer",
        "answer_id": "5",
        "at": "2020-01-01T00:52:00Z",
        "attempt_id": "1",
        "from_answer_id": "5",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 100,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "4",
        "at": "2020-01-01T00:51:00Z",
        "attempt_id": "1",
        "from_answer_id": "4",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 100,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "current_answer",
        "answer_id": "25",
        "at": "2020-01-01T00:50:00Z",
        "attempt_id": "1",
        "from_answer_id": "25",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "saved_answer",
        "answer_id": "24",
        "at": "2020-01-01T00:49:00Z",
        "attempt_id": "1",
        "from_answer_id": "24",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "23",
        "at": "2020-01-01T00:48:00Z",
        "attempt_id": "1",
        "from_answer_id": "23",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "28",
        "at": "2020-01-01T00:47:00Z",
        "attempt_id": "1",
        "from_answer_id": "28",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "submission",
        "answer_id": "12",
        "at": "2020-01-01T00:46:00Z",
        "attempt_id": "1",
        "from_answer_id": "12",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "16",
        "at": "2020-01-01T00:45:00Z",
        "attempt_id": "1",
        "from_answer_id": "16",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:44:00Z",
        "attempt_id": "0",
        "from_answer_id": "16",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "11",
        "at": "2020-01-01T00:43:00Z",
        "attempt_id": "0",
        "from_answer_id": "11",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "17",
        "at": "2020-01-01T00:42:00Z",
        "attempt_id": "0",
        "from_answer_id": "17",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:41:00Z",
        "attempt_id": "0",
        "from_answer_id": "17",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:40:00Z",
        "attempt_id": "1",
        "from_answer_id": "17",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "1",
        "at": "2020-01-01T00:39:00Z",
        "attempt_id": "0",
        "from_answer_id": "1",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 99,
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "7",
        "at": "2020-01-01T00:38:00Z",
        "attempt_id": "0",
        "from_answer_id": "7",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "score": 98,
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "submission",
        "answer_id": "22",
        "at": "2020-01-01T00:37:00Z",
        "attempt_id": "1",
        "from_answer_id": "22",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "26",
        "at": "2020-01-01T00:36:00Z",
        "attempt_id": "1",
        "from_answer_id": "26",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "submission",
        "answer_id": "21",
        "at": "2020-01-01T00:35:00Z",
        "attempt_id": "0",
        "from_answer_id": "21",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "submission",
        "answer_id": "27",
        "at": "2020-01-01T00:34:00Z",
        "attempt_id": "0",
        "from_answer_id": "27",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"id": "31", "login": "jane"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:33:00Z",
        "attempt_id": "1",
        "from_answer_id": "27",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:32:00Z",
        "attempt_id": "0",
        "from_answer_id": "27",
        "item": {"id": "201", "string": {"title": "Chapitre 1"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:31:00Z",
        "attempt_id": "1",
        "from_answer_id": "27",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:30:00Z",
        "attempt_id": "0",
        "from_answer_id": "27",
        "item": {"id": "202", "string": {"title": "Chapitre 2"}, "type": "Chapter"},
        "participant": {"id": "11", "name": "user", "type": "User"},
        "user": {"first_name": "John", "id": "11", "last_name": "Doe", "login": "user"}
      },
      {
        "activity_type": "result_validated",
        "at": "2020-01-01T00:29:00Z",
        "attempt_id": "1",
        "from_answer_id": "27",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
        "participant": {"id": "41", "name": "paul", "type": "User"},
        "user": {"first_name": "Paul", "id": "41", "last_name": "Smith", "login": "paul"}
      },
      {
        "activity_type": "result_started",
        "at": "2020-01-01T00:28:00Z",
        "attempt_id": "1",
        "from_answer_id": "27",
        "item": {"id": "200", "string": {"title": "Tache 1"}, "type": "Task"},
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
        "at": "2017-05-30T12:00:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "from_answer_id": "-1"
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
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
        "at": "2017-05-30T12:00:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
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
        "at": "2017-05-29T06:38:00Z",
        "participant": {"id": "31", "name": "jane", "type": "User"},
        "attempt_id": "1",
        "item": {"id": "200", "string": {"title": "Task 1"}, "type": "Task"},
        "user": {"id": "31", "first_name": "Jane", "last_name": "Doe", "login": "jane"},
        "from_answer_id": "-1"
      }
    ]
    """
  Examples:
    | forceStraightJoinInItemActivityLog |
    | force                              |
    | no                                 |
