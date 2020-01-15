Feature: Get user's answer by user_answer_id
Background:
  Given the database has the following table 'groups':
    | id | name    | type     |
    | 11 | jdoe    | UserSelf |
    | 13 | Group B | Class    |
    | 21 | other   | UserSelf |
    | 23 | Group C | Class    |
  And the database has the following table 'users':
    | login | temp_user | group_id | first_name | last_name |
    | jdoe  | 0         | 11       | John       | Doe       |
    | other | 0         | 21       | George     | Bush      |
  And the database has the following table 'groups_groups':
    | id | parent_group_id | child_group_id |
    | 61 | 13              | 11             |
    | 62 | 13              | 21             |
    | 63 | 23              | 21             |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
    | 73 | 13                | 13             | 1       |
    | 74 | 13                | 11             | 0       |
    | 75 | 13                | 21             | 0       |
    | 76 | 23                | 21             | 0       |
    | 77 | 23                | 23             | 1       |
  And the database has the following table 'items':
    | id  |
    | 200 |
    | 210 |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 200     | content                  |
    | 23       | 210     | content_with_descendants |
  And the database has the following table 'attempts':
    | id  | group_id | item_id | order |
    | 150 | 11       | 200     | 0     |
    | 250 | 13       | 210     | 0     |
  And the database has the following table 'answers':
    | id  | author_id | attempt_id | type       | state   | answer   | created_at          |
    | 101 | 11        | 150        | Submission | Current | print(1) | 2017-05-29 06:38:38 |
    | 102 | 11        | 250        | Submission | Current | print(2) | 2017-05-29 06:38:38 |
  And the database has the following table 'gradings':
    | answer_id | score | graded_at           |
    | 101       | 100   | 2018-05-29 06:38:38 |
    | 102       | 100   | 2019-05-29 06:38:38 |

  Scenario: User has access to the item and the answers.author_id = authenticated user's self group
    Given I am the user with id "11"
    When I send a GET request to "/answers/101"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "101",
      "attempt_id": "150",
      "score": 100.0,
      "answer": "print(1)",
      "state": "Current",
      "created_at": "2017-05-29T06:38:38Z",
      "type": "Submission",
      "item_id": "200",
      "author_id": "11",
      "graded_at": "2018-05-29T06:38:38Z"
    }
    """

  Scenario: User has access to the item and the user is a team member of attempts.group_id
    Given I am the user with id "21"
    When I send a GET request to "/answers/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "102",
      "attempt_id": "250",
      "score": 100,
      "answer": "print(2)",
      "state": "Current",
      "created_at": "2017-05-29T06:38:38Z",
      "type": "Submission",
      "item_id": "210",
      "author_id": "11",
      "graded_at": "2019-05-29T06:38:38Z"
    }
    """
