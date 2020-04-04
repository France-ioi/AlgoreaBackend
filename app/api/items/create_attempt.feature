Feature: Create an attempt for an item
  Background:
    Given the database has the following table 'groups':
      | id  | type |
      | 101 | User |
      | 102 | Team |
      | 111 | User |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
      | jane  | 111      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 102             | 101            |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 101               | 101            |
      | 102               | 101            |
      | 102               | 102            |
      | 111               | 111            |
    And the database has the following table 'items':
      | id | url                                                                     | type    | allows_multiple_attempts | default_language_tag |
      | 10 | null                                                                    | Chapter | 0                        | fr                   |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 0                        | fr                   |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 1                        | fr                   |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 101      | 50      | content                  |
      | 102      | 60      | content                  |
      | 111      | 10      | content_with_descendants |
      | 111      | 50      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
      | 0  | 102            |
      | 0  | 111            |

  Scenario Outline: User is able to create an attempt for his self group
    Given I am the user with id "111"
    When I send a POST request to "/items/<item_id>/attempts"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "0"
      }
      """
    And the table "attempts" should stay unchanged
    And the table "results" should be:
      | attempt_id | participant_id | item_id   | score_computed | tasks_tried | result_propagation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | latest_submission_at | score_obtained_at | validated_at | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 0          | 111            | <item_id> | 0              | 0           | done                     | 1                                                         | null                 | null              | null         | 1                                                 |
  Examples:
    | item_id |
    | 50      |
    | 10      |

  Scenario: User is able to create an attempt as a team
    Given I am the user with id "101"
    When I send a POST request to "/items/60/attempts?as_team_id=102"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "0"
      }
      """
    And the table "attempts" should stay unchanged
    And the table "results" should be:
      | attempt_id | participant_id | item_id | score_computed | tasks_tried | result_propagation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | latest_submission_at | score_obtained_at | validated_at | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 0          | 102            | 60      | 0              | 0           | done                     | 1                                                         | null                 | null              | null         | 1                                                 |
    When I send a POST request to "/items/60/attempts?as_team_id=102"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "1"
      }
      """
    And the table "attempts" should be:
      | id | participant_id |
      | 0  | 101            |
      | 0  | 102            |
      | 0  | 111            |
      | 1  | 102            |
    And the table "results" should be:
      | attempt_id | participant_id | item_id | score_computed | tasks_tried | result_propagation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | latest_submission_at | score_obtained_at | validated_at | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 0          | 102            | 60      | 0              | 0           | done                     | 1                                                         | null                 | null              | null         | 1                                                 |
      | 1          | 102            | 60      | 0              | 0           | done                     | 1                                                         | null                 | null              | null         | 1                                                 |
