Feature: Update the 'current' answer
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database table 'groups' has also the following row:
      | id | type |
      | 13 | Team |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 101               | 101            |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 101            |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 50 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 1  | 13             |
      | 1  | 101            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id |
      | 1          | 13             | 50      |
      | 1          | 101            | 50      |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | type       | created_at          |
      | 100 | 101       | 101            | 1          | 50      | Submission | 2017-05-29 06:38:38 |

  Scenario: User is able to create the 'current' answer
    Given I am the user with id "101"
    When I send a PUT request to "/items/50/attempts/1/answers/current" with the following body:
      """
      {
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
    And the table "answers" should be:
      | author_id | participant_id | attempt_id | item_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 101            | 1          | 50      | Submission | null    | null       | 0                                                 |
      | 101       | 101            | 1          | 50      | Current    | print 1 | some state | 1                                                 |

  Scenario: User is able to create the 'current' answer for a team attempt
    Given I am the user with id "101"
    When I send a PUT request to "/items/50/attempts/1/answers/current?as_team_id=13" with the following body:
      """
      {
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
    And the table "answers" should be:
      | author_id | participant_id | attempt_id | item_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 13             | 1          | 50      | Current    | print 1 | some state | 1                                                 |
      | 101       | 101            | 1          | 50      | Submission | null    | null       | 0                                                 |

  Scenario: User is able to replace the 'current' answer
    Given I am the user with id "101"
    And the database has the following table 'answers':
      | author_id | participant_id | attempt_id | item_id | type    | created_at          |
      | 101       | 101            | 1          | 50      | Current | 2017-05-29 06:38:38 |
    When I send a PUT request to "/items/50/attempts/1/answers/current" with the following body:
      """
      {
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
    And the table "answers" should be:
      | author_id | participant_id | attempt_id | type       | answer  | state      | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 101            | 1          | Submission | null    | null       | 0                                                 |
      | 101       | 101            | 1          | Current    | print 1 | some state | 1                                                 |
