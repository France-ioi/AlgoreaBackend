Feature: Get a current answer
Background:
  Given the database has the following table "groups":
    | id  | name         | type  |
    | 13  | Team         | Team  |
    | 14  | Group B      | Class |
    | 15  | TeamNoAnswer | Team     |
    | 23  | Group C      | Class |
  And the database has the following users:
    | group_id | login        | first_name   | last_name |
    | 11       | jdoe         | John         | Doe       |
    | 12       | jdoenoanswer | JohnNoAnswer | Doe       |
    | 21       | manager      | Man          | Ager      |
    | 100      | top          | Top          | Score     |
  And the database has the following table "groups_groups":
    | parent_group_id | child_group_id |
    | 14              | 11             |
    | 13              | 21             |
    | 15              | 21             |
    | 23              | 21             |
  And the groups ancestors are computed
  And the database has the following table "group_managers":
    | group_id | manager_id | can_watch_members |
    | 13       | 21         | true              |
    | 15       | 21         | true              |
  And the groups ancestors are computed
  And the database has the following table "items":
    | id  | default_language_tag |
    | 200 | fr                   |
    | 210 | fr                   |
  And the database has the following table "permissions_generated":
    | group_id | item_id | can_view_generated       | can_watch_generated |
    | 12       | 200     | content                  | answer              |
    | 13       | 200     | content                  | answer              |
    | 14       | 200     | content                  | none                |
    | 15       | 200     | content                  | answer              |
    | 23       | 210     | content_with_descendants | answer              |
  And the database has the following table "results":
    | attempt_id | participant_id | item_id |
    | 1          | 11             | 200     |
    | 2          | 11             | 200     |
    | 3          | 11             | 200     |
    | 1          | 13             | 200     |
    | 1          | 100            | 200     |
    | 1          | 13             | 210     |
    | 1          | 11             | 210     |
    | 1          | 100            | 210     |
  And the database has the following table "answers":
    | id  | author_id | participant_id | attempt_id | item_id | type       | state    | answer    | created_at          |
    | 101 | 11        | 11             | 1          | 200     | Submission | State101 | print(1)  | 2020-01-01 06:00:00 |
    | 102 | 11        | 11             | 2          | 200     | Submission | State102 | print(2)  | 2020-01-01 07:00:00 |
    | 103 | 11        | 11             | 3          | 200     | Submission | State103 | print(3)  | 2020-01-01 08:00:00 |
    | 104 | 11        | 11             | 3          | 200     | Submission | State104 | print(4)  | 2020-01-01 09:00:00 |
    | 105 | 11        | 13             | 1          | 200     | Submission | State105 | print(5)  | 2020-01-01 10:00:00 |
    | 106 | 11        | 13             | 1          | 210     | Submission | State106 | print(6)  | 2020-01-01 11:00:00 |
    | 107 | 11        | 13             | 1          | 210     | Submission | State107 | print(7)  | 2020-01-01 12:00:00 |
    | 108 | 11        | 13             | 1          | 210     | Submission | State108 | print(8)  | 2020-01-01 13:00:00 |
    | 109 | 11        | 11             | 1          | 210     | Submission | State109 | print(9)  | 2020-01-01 14:00:00 |
    | 110 | 100       | 100            | 1          | 200     | Submission | State110 | print(10) | 2020-01-01 15:00:00 |
    | 111 | 100       | 100            | 1          | 210     | Submission | State111 | print(11) | 2020-01-01 16:00:00 |
  And the database has the following table "gradings":
    | answer_id | score | graded_at           |
    | 101       | 91    | 2020-01-01 06:00:01 |
    | 102       | 97    | 2020-01-01 07:00:01 |
    | 103       | 97    | 2020-01-01 08:00:01 |
    | 104       | 96    | 2020-01-01 09:00:01 |
    | 105       | 99    | 2020-01-01 10:00:01 |
    | 106       | 98    | 2020-01-01 11:00:01 |
    | 107       | 98    | 2020-01-01 12:00:01 |
    | 108       | 96    | 2020-01-01 13:00:01 |
    | 109       | 96    | 2020-01-01 14:00:01 |
    | 110       | 100   | 2020-01-01 15:00:01 |
    | 111       | 100   | 2020-01-01 16:00:01 |

  Scenario: User has access to the item and retrieves his best answer
    Given I am the user with id "11"
    When I send a GET request to "/items/200/best-answer"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "103",
      "attempt_id": "3",
      "participant_id": "11",
      "score": 97.0,
      "answer": "print(3)",
      "state": "State103",
      "created_at": "2020-01-01T08:00:00Z",
      "type": "Submission",
      "item_id": "200",
      "author_id": "11",
      "graded_at": "2020-01-01T08:00:01Z"
    }
    """

  Scenario: Should return a 404 when user has access to the item but has no answer
    Given I am the user with id "12"
    When I send a GET request to "/items/200/best-answer"
    Then the response code should be 404

  Scenario: User has access to the item and retrieves the best answer given by watched_group_id
    Given I am the user with id "21"
    When I send a GET request to "/items/210/best-answer?watched_group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "107",
      "attempt_id": "1",
      "participant_id": "13",
      "score": 98.0,
      "answer": "print(7)",
      "state": "State107",
      "created_at": "2020-01-01T12:00:00Z",
      "type": "Submission",
      "item_id": "210",
      "author_id": "11",
      "graded_at": "2020-01-01T12:00:01Z"
    }
    """

  Scenario: Should return a 404 when user has access to the item but the watched_group_id has no answer
    Given I am the user with id "21"
    When I send a GET request to "/items/210/best-answer?watched_group_id=15"
    Then the response code should be 404
