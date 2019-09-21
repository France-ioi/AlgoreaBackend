Feature: Update the 'current' answer
  Background:
    Given the database has the following table 'users':
      | id  | login | self_group_id |
      | 10  | john  | 101           |
    And the database has the following table 'groups':
      | id  |
      | 101 |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type   | status_date |
      | 15 | 22              | 13             | direct | null        |
    And the database has the following table 'items':
      | id |
      | 50 |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_date | creator_user_id |
      | 101      | 50      | 2017-05-29 06:38:38        | 10              |
    And the database has the following table 'users_answers':
      | id  | user_id | item_id | attempt_id | type       | submission_date     |
      | 100 | 10      | 50      | 200        | Submission | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 200 | 101      | 50      | 0     |

  Scenario: User is able to create the 'current' answer and users_items.active_attempt_id = request.attempt_id
    Given I am the user with id "10"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 10      | 50      | 200               |
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
      | user_id | item_id | active_attempt_id | answer  | state      |
      | 10      | 50      | 200               | print 1 | some state |
    And the table "users_answers" should be:
      | user_id | item_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, submission_date, NOW())) < 3 |
      | 10      | 50      | 200        | Submission | null    | null       | 0                                                      |
      | 10      | 50      | 200        | Current    | print 1 | some state | 1                                                      |

  Scenario: User is able to create the 'current' answer and users_items.active_attempt_id != request.attempt_id
    Given I am the user with id "10"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id | answer | state |
      | 10      | 50      | 100               | null   | null  |
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
    And the table "users_answers" should be:
      | user_id | item_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, submission_date, NOW())) < 3 |
      | 10      | 50      | 200        | Submission | null    | null       | 0                                                      |
      | 10      | 50      | 200        | Current    | print 1 | some state | 1                                                      |

  Scenario: User is able to update the 'current' answer
    Given I am the user with id "10"
    And the database has the following table 'users_answers':
      | id  | user_id | item_id | attempt_id | type    | submission_date     |
      | 101 | 10      | 50      | 200        | Current | 2017-05-29 06:38:38 |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 10      | 50      | 200               |
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
      | user_id | item_id | active_attempt_id | answer  | state      |
      | 10      | 50      | 200               | print 1 | some state |
    And the table "users_answers" should be:
      | id  | user_id | item_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, submission_date, NOW())) < 3 |
      | 100 | 10      | 50      | 200        | Submission | null    | null       | 0                                                      |
      | 101 | 10      | 50      | 200        | Current    | print 1 | some state | 0                                                      |
