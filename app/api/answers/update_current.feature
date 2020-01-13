Feature: Update the 'current' answer
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 15 | 22              | 13             |
    And the database has the following table 'items':
      | id |
      | 50 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 101      | 50      | 0     |
      | 200 | 101      | 50      | 0     |
    And the database has the following table 'answers':
      | id  | author_id | attempt_id | type       | created_at          |
      | 100 | 101       | 200        | Submission | 2017-05-29 06:38:38 |

  Scenario: User is able to create the 'current' answer and users_items.active_attempt_id = request.attempt_id
    Given I am the user with id "101"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 200               |
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "200",
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 200               |
    And the table "answers" should be:
      | author_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 200        | Submission | null    | null       | 0                                                 |
      | 101       | 200        | Current    | print 1 | some state | 1                                                 |

  Scenario: User is able to create the 'current' answer and users_items.active_attempt_id != request.attempt_id
    Given I am the user with id "101"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 100               |
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "200",
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should stay unchanged
    And the table "answers" should be:
      | author_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 200        | Submission | null    | null       | 0                                                 |
      | 101       | 200        | Current    | print 1 | some state | 1                                                 |

  Scenario: User is able to update the 'current' answer
    Given I am the user with id "101"
    And the database has the following table 'answers':
      | id  | author_id | attempt_id | type    | created_at          |
      | 101 | 101       | 200        | Current | 2017-05-29 06:38:38 |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 200               |
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "200",
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 200               |
    And the table "answers" should be:
      | id  | author_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 100 | 101       | 200        | Submission | null    | null       | 0                                                 |
      | 101 | 101       | 200        | Current    | print 1 | some state | 0                                                 |
