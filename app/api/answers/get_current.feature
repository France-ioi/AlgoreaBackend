Feature: Get a current answer
Background:
  Given the database has the following table 'groups':
    | id | name    | type  |
    | 11 | jdoe    | User  |
    | 13 | Team    | Team  |
    | 14 | Group B | Class |
    | 21 | other   | User  |
    | 23 | Group C | Class |
  And the database has the following table 'users':
    | login | temp_user | group_id | first_name | last_name |
    | jdoe  | 0         | 11       | John       | Doe       |
    | other | 0         | 21       | George     | Bush      |
  And the database has the following table 'groups_groups':
    | parent_group_id | child_group_id |
    | 14              | 11             |
    | 13              | 21             |
    | 23              | 21             |
  And the groups ancestors are computed
  And the database has the following table 'items':
    | id  | default_language_tag |
    | 200 | fr                   |
    | 210 | fr                   |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 14       | 200     | content                  |
    | 23       | 210     | content_with_descendants |
  And the database has the following table 'results':
    | attempt_id | participant_id | item_id |
    | 1          | 11             | 200     |
    | 1          | 13             | 210     |
  And the database has the following table 'answers':
    | id  | author_id | participant_id | attempt_id | item_id | type       | state   | answer   | created_at          |
    |   1 | 11        | 11             | 1          | 200     | Submission | Current | print(1) | 2017-05-29 06:38:37 |
    |   2 | 11        | 11             | 1          | 200     | Saved      | Current | print(2) | 2017-05-29 06:38:37 |
    |   3 | 11        | 11             | 1          | 200     | Current    | Current | print(3) | 2017-05-29 06:38:37 |
    |   4 | 11        | 13             | 1          | 210     | Submission | Current | print(4) | 2017-05-29 06:38:37 |
    |   5 | 11        | 13             | 1          | 210     | Saved      | Current | print(5) | 2017-05-29 06:38:37 |
    |   6 | 11        | 13             | 1          | 210     | Current    | Current | print(6) | 2017-05-29 06:38:37 |
    | 101 | 11        | 11             | 1          | 200     | Submission | Current | print(1) | 2017-05-29 06:38:38 |
    | 102 | 11        | 11             | 1          | 200     | Saved      | Current | print(2) | 2017-05-29 06:38:38 |
    | 103 | 11        | 11             | 1          | 200     | Current    | Current | print(3) | 2017-05-29 06:38:38 |
    | 104 | 11        | 13             | 1          | 210     | Submission | Current | print(4) | 2017-05-29 06:38:38 |
    | 105 | 11        | 13             | 1          | 210     | Saved      | Current | print(5) | 2017-05-29 06:38:38 |
    | 106 | 11        | 13             | 1          | 210     | Current    | Current | print(6) | 2017-05-29 06:38:38 |
    | 201 | 11        | 11             | 1          | 200     | Submission | Current | print(1) | 2017-05-29 06:38:38 |
    | 202 | 11        | 11             | 1          | 200     | Saved      | Current | print(2) | 2017-05-29 06:38:38 |
    | 203 | 11        | 11             | 1          | 200     | Current    | Current | print(3) | 2017-05-29 06:38:38 |
    | 204 | 11        | 13             | 1          | 210     | Submission | Current | print(4) | 2017-05-29 06:38:38 |
    | 205 | 11        | 13             | 1          | 210     | Saved      | Current | print(5) | 2017-05-29 06:38:38 |
    | 206 | 11        | 13             | 2          | 210     | Current    | Current | print(6) | 2017-05-29 06:38:38 |
    | 207 | 11        | 11             | 2          | 200     | Current    | Current | print(7) | 2018-05-29 06:38:37 |
  And the database has the following table 'gradings':
    | answer_id | score | graded_at           |
    | 101       | 91    | 2018-05-29 06:38:31 |
    | 102       | 92    | 2019-05-29 06:38:32 |
    | 103       | 93    | 2018-05-29 06:38:33 |
    | 104       | 94    | 2019-05-29 06:38:34 |
    | 105       | 95    | 2018-05-29 06:38:35 |
    | 106       | 96    | 2019-05-29 06:38:36 |

  Scenario: User has access to the item and the answers.participant_id = authenticated user's self group
    Given I am the user with id "11"
    When I send a GET request to "/items/200/current-answer?attempt_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "103",
      "attempt_id": "1",
      "participant_id": "11",
      "score": 93.0,
      "answer": "print(3)",
      "state": "Current",
      "created_at": "2017-05-29T06:38:38Z",
      "type": "Current",
      "item_id": "200",
      "author_id": "11",
      "graded_at": "2018-05-29T06:38:33Z"
    }
    """

  Scenario: User has access to the item and the user is a team member of attempts.participant_id
    Given I am the user with id "21"
    When I send a GET request to "/items/210/current-answer?as_team_id=13&attempt_id=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "206",
      "attempt_id": "2",
      "participant_id": "13",
      "score": null,
      "answer": "print(6)",
      "state": "Current",
      "created_at": "2017-05-29T06:38:38Z",
      "type": "Current",
      "item_id": "210",
      "author_id": "11",
      "graded_at": null
    }
    """
