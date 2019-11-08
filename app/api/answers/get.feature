Feature: Get user's answer by user_answer_id
Background:
  Given the database has the following table 'groups':
    | id | name        | type      |
    | 11 | jdoe        | UserSelf  |
    | 12 | jdoe-admin  | UserAdmin |
    | 13 | Group B     | Class     |
    | 21 | other       | UserSelf  |
    | 22 | other-admin | UserAdmin |
    | 23 | Group C     | Class     |
  And the database has the following table 'users':
    | login | temp_user | group_id | owned_group_id | first_name | last_name |
    | jdoe  | 0         | 11       | 12             | John       | Doe       |
    | other | 0         | 21       | 22             | George     | Bush      |
  And the database has the following table 'groups_groups':
    | id | parent_group_id | child_group_id | type               |
    | 61 | 13              | 11             | invitationAccepted |
    | 62 | 13              | 21             | requestAccepted    |
    | 63 | 23              | 21             | direct             |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
    | 72 | 12                | 12             | 1       |
    | 73 | 13                | 13             | 1       |
    | 74 | 13                | 11             | 0       |
    | 75 | 13                | 21             | 0       |
    | 76 | 23                | 21             | 0       |
    | 77 | 23                | 23             | 1       |
  And the database has the following table 'items':
    | id  | has_attempts |
    | 200 | 0            |
    | 210 | 1            |
  And the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_since | cached_partial_access_since |
    | 43 | 13       | 200     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         |
    | 46 | 23       | 210     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         |
  And the database has the following table 'users_answers':
    | id  | user_id | item_id | attempt_id | type       | state   | answer   | lang_prog | submitted_at        | score | validated | graded_at           |
    | 101 | 11      | 200     | 150        | Submission | Current | print(1) | python    | 2017-05-29 06:38:38 | 100   | true      | 2018-05-29 06:38:38 |
    | 102 | 11      | 210     | 250        | Submission | Current | print(2) | python    | 2017-05-29 06:38:38 | 100   | true      | 2019-05-29 06:38:38 |
  And the database has the following table 'groups_attempts':
    | id  | group_id | item_id | order |
    | 150 | 11       | 200     | 0     |
    | 250 | 13       | 210     | 0     |

  Scenario: User has access to the item and the users_answers.user_id = authenticated user's self group
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
      "submitted_at": "2017-05-29T06:38:38Z",
      "type": "Submission",
      "item_id": "200",
      "user_id": "11",
      "graded_at": "2018-05-29T06:38:38Z",
      "validated": true
    }
    """

  Scenario: User has access to the item and the user is a team member of groups_attempts.group_id (items.has_attempts=1)
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
      "submitted_at": "2017-05-29T06:38:38Z",
      "type": "Submission",
      "item_id": "210",
      "user_id": "11",
      "graded_at": "2019-05-29T06:38:38Z",
      "validated": true
    }
    """
