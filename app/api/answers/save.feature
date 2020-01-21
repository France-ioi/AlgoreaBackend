Feature: Save an answer
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 15 | 13              | 101            |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 50 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
    And the database has the following table 'attempts':
      | id  | group_id | item_id | order |
      | 100 | 13       | 50      | 1     |
      | 200 | 101      | 50      | 1     |
    And the database has the following table 'answers':
      | id  | author_id | attempt_id | type       | created_at          |
      | 100 | 101       | 200        | Submission | 2017-05-29 06:38:38 |

  Scenario: User is able to save an answer
    Given I am the user with id "101"
    When I send a POST request to "/attempts/200/answers" with the following body:
      """
      {
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should be:
      | author_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 200        | Submission | null    | null       | 0                                                 |
      | 101       | 200        | Saved      | print 1 | some state | 1                                                 |

  Scenario: User is able to save an answer for a team attempt
    Given I am the user with id "101"
    When I send a POST request to "/attempts/100/answers" with the following body:
      """
      {
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should be:
      | author_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 100        | Saved      | print 1 | some state | 1                                                 |
      | 101       | 200        | Submission | null    | null       | 0                                                 |
