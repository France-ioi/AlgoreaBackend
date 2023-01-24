Feature: Get a current answer
Background:
  Given the database has the following table 'groups':
    | id | name    | type  |
    | 11 | jdoe    | User  |
    | 13 | Team    | Team  |
    | 14 | Group B | Class |
    | 21 | manager | User  |
    | 23 | Group C | Class |
  And the database has the following table 'users':
    | login   | group_id | first_name | last_name |
    | jdoe    | 11       | John       | Doe       |
    | manager | 21       | Man        | Ager      |
  And the database has the following table 'groups_groups':
    | parent_group_id | child_group_id |
    | 14              | 11             |
    | 13              | 21             |
    | 23              | 21             |
  And the groups ancestors are computed
  And the database has the following table 'group_managers':
    | group_id | manager_id | can_watch_members |
    | 13       | 21         | true              |
  And the groups ancestors are computed
  And the database has the following table 'items':
    | id  | default_language_tag |
    | 200 | fr                   |
    | 210 | fr                   |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       | can_watch_generated |
    | 13       | 200     | content                  | answer              |
    | 14       | 200     | content                  | none                |
    | 23       | 210     | content_with_descendants | answer                |
  And the database has the following table 'results':
    | attempt_id | participant_id | item_id |
    | 1          | 11             | 200     |
    | 2          | 11             | 200     |
    | 3          | 11             | 200     |
    | 1          | 13             | 210     |
    | 1          | 11             | 210     |
  And the database has the following table 'answers':
    | id  | author_id | participant_id | attempt_id | item_id | type       | state    | answer   | created_at          |
    | 101 | 11        | 11             | 1          | 200     | Submission | State101 | print(1) | 2017-05-29 06:38:38 |
    | 102 | 11        | 11             | 2          | 200     | Submission | State102 | print(2) | 2017-05-29 07:38:38 |
    | 103 | 11        | 11             | 3          | 200     | Submission | State103 | print(3) | 2017-05-29 08:38:38 |
    | 104 | 11        | 13             | 1          | 210     | Submission | State104 | print(4) | 2017-05-29 09:38:38 |
    | 105 | 11        | 13             | 1          | 210     | Submission | State105 | print(5) | 2017-05-29 10:38:38 |
    | 106 | 11        | 13             | 1          | 210     | Submission | State106 | print(6) | 2017-05-29 11:38:38 |
    | 107 | 11        | 11             | 1          | 210     | Submission | State107 | print(7) | 2017-05-29 08:38:38 |
  And the database has the following table 'gradings':
    | answer_id | score | graded_at           |
    | 101       | 91    | 2018-05-29 06:38:31 |
    | 102       | 97    | 2019-05-29 06:38:32 |
    | 103       | 97    | 2018-05-29 06:38:33 |
    | 104       | 96    | 2019-05-29 06:38:34 |
    | 105       | 96    | 2018-05-29 06:38:35 |
    | 106       | 95    | 2019-05-29 06:38:36 |
    | 107       | 98    | 2019-05-29 06:38:38 |

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
      "created_at": "2017-05-29T08:38:38Z",
      "type": "Submission",
      "item_id": "200",
      "author_id": "11",
      "graded_at": "2018-05-29T06:38:33Z"
    }
    """

  Scenario: User has access to the item and retrieves the best answer given by watched_group_id
    Given I am the user with id "21"
    When I send a GET request to "/items/210/best-answer?watched_group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "105",
      "attempt_id": "1",
      "participant_id": "13",
      "score": 96.0,
      "answer": "print(5)",
      "state": "State105",
      "created_at": "2017-05-29T10:38:38Z",
      "type": "Submission",
      "item_id": "210",
      "author_id": "11",
      "graded_at": "2018-05-29T06:38:35Z"
    }
    """
