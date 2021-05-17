Feature: Create an attempt for an item
  Background:
    Given the database has the following table 'groups':
      | id  | type  | root_activity_id |
      | 99  | Other | 50               |
      | 100 | Class | 10               |
      | 101 | User  | null             |
      | 102 | Team  | 60               |
      | 111 | User  | null             |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
      | jane  | 111      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 99              | 111            |
      | 100             | 111            |
      | 102             | 101            |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id | url                                                                     | type    | allows_multiple_attempts | default_language_tag |
      | 10 | null                                                                    | Chapter | 1                        | fr                   |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 1                        | fr                   |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 1                        | fr                   |
      | 70 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 1                        | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 1           |
      | 60             | 70            | 1           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
      | 10               | 70            |
      | 60               | 70            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 101      | 50      | content                  |
      | 102      | 10      | content                  |
      | 102      | 60      | content                  |
      | 102      | 70      | content                  |
      | 111      | 10      | content_with_descendants |
      | 111      | 50      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          |
      | 0  | 101            | 2019-05-30 11:00:00 |
      | 0  | 102            | 2019-05-30 11:00:00 |
      | 0  | 111            | 2019-05-30 11:00:00 |

  Scenario Outline: User is able to create an attempt for his self group
    Given I am the user with id "111"
    When I send a POST request to "/items/<item_id>/attempts?parent_attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"id": "1"}
      }
      """
    And the table "attempts" should be:
      | id | participant_id | root_item_id | parent_attempt_id | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 0  | 101            | null         | null              | 0                                                 |
      | 0  | 102            | null         | null              | 0                                                 |
      | 0  | 111            | null         | null              | 0                                                 |
      | 1  | 111            | <item_id>    | 0                 | 1                                                 |
    And the table "results" should be:
      | attempt_id | participant_id | item_id   | score_computed | tasks_tried | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | latest_submission_at | score_obtained_at | validated_at | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 1          | 111            | <item_id> | 0              | 0           | 1                                                         | null                 | null              | null         | 1                                                 |
    And the table "results_propagate" should be empty
  Examples:
    | item_id |
    | 50      |
    | 10      |

  Scenario: User is able to create an attempt as a team
    Given I am the user with id "101"
    When I send a POST request to "/items/60/attempts?as_team_id=102&parent_attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"id": "1"}
      }
      """
    And the table "attempts" should be:
      | id | participant_id | root_item_id | parent_attempt_id | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 0  | 101            | null         | null              | 0                                                 |
      | 0  | 102            | null         | null              | 0                                                 |
      | 0  | 111            | null         | null              | 0                                                 |
      | 1  | 102            | 60           | 0                 | 1                                                 |
    And the table "results" should be:
      | attempt_id | participant_id | item_id | score_computed | tasks_tried | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | latest_submission_at | score_obtained_at | validated_at | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 0          | 102            | 10      | 0              | 0           | 1                                                         | null                 | null              | null         | null                                              |
      | 1          | 102            | 60      | 0              | 0           | 1                                                         | null                 | null              | null         | 1                                                 |
    And the table "results_propagate" should be empty
    When I send a POST request to "/items/60/attempts?as_team_id=102&parent_attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"id": "2"}
      }
      """
    And the table "attempts" should be:
      | id | participant_id | root_item_id | parent_attempt_id | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 0  | 101            | null         | null              | 0                                                 |
      | 0  | 102            | null         | null              | 0                                                 |
      | 0  | 111            | null         | null              | 0                                                 |
      | 1  | 102            | 60           | 0                 | 1                                                 |
      | 2  | 102            | 60           | 0                 | 1                                                 |
    And the table "results" should be:
      | attempt_id | participant_id | item_id | score_computed | tasks_tried | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | latest_submission_at | score_obtained_at | validated_at | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 0          | 102            | 10      | 0              | 0           | 1                                                         | null                 | null              | null         | null                                              |
      | 1          | 102            | 60      | 0              | 0           | 1                                                         | null                 | null              | null         | 1                                                 |
      | 2          | 102            | 60      | 0              | 0           | 1                                                         | null                 | null              | null         | 1                                                 |
    And the table "results_propagate" should be empty
    When I send a POST request to "/items/60/70/attempts?as_team_id=102&parent_attempt_id=2"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"id": "3"}
      }
      """
    And the table "attempts" should be:
      | id | participant_id | root_item_id | parent_attempt_id | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 0  | 101            | null         | null              | 0                                                 |
      | 0  | 102            | null         | null              | 0                                                 |
      | 0  | 111            | null         | null              | 0                                                 |
      | 1  | 102            | 60           | 0                 | 1                                                 |
      | 2  | 102            | 60           | 0                 | 1                                                 |
      | 3  | 102            | 70           | 2                 | 1                                                 |
    And the table "results" should be:
      | attempt_id | participant_id | item_id | score_computed | tasks_tried | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | latest_submission_at | score_obtained_at | validated_at | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 0          | 102            | 10      | 0              | 0           | 1                                                         | null                 | null              | null         | null                                              |
      | 1          | 102            | 60      | 0              | 0           | 1                                                         | null                 | null              | null         | 1                                                 |
      | 2          | 102            | 60      | 0              | 0           | 1                                                         | null                 | null              | null         | 1                                                 |
      | 3          | 102            | 70      | 0              | 0           | 1                                                         | null                 | null              | null         | 1                                                 |
    And the table "results_propagate" should be empty
