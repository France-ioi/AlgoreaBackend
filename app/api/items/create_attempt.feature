Feature: Create an attempt for an item
  Background:
    Given the database has the following table 'groups':
      | id  | team_item_id | type     |
      | 101 | null         | UserSelf |
      | 102 | 10           | Team     |
      | 111 | null         | UserSelf |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
      | jane  | 111      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 102             | 101            |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
      | 102               | 101            | 0       |
      | 102               | 102            | 1       |
      | 111               | 111            | 1       |
    And the database has the following table 'items':
      | id | url                                                                     | type    | has_attempts |
      | 10 | null                                                                    | Chapter | 0            |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 0            |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 1            |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 101      | 50      | content                  |
      | 102      | 60      | content                  |
      | 111      | 50      | content_with_descendants |

  Scenario: User is able to create an attempt for his self group
    Given I am the user with id "111"
    When I send a POST request to "/items/50/attempts"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "5577006791947779410"
      }
      """
    And the table "attempts" should be:
      | id                  | group_id | item_id | score_computed | tasks_tried | result_propagation_state | latest_activity_at | latest_answer_at | score_obtained_at | validated_at | started_at |
      | 5577006791947779410 | 111      | 50      | 0              | 0           | done                     | null               | null             | null              | null         | null       |

  Scenario: User is able to create an attempt as a team
    Given I am the user with id "101"
    When I send a POST request to "/items/60/attempts?as_team_id=102"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "5577006791947779410"
      }
      """
    And the table "attempts" should be:
      | id                  | group_id | item_id | score_computed | tasks_tried | result_propagation_state | latest_activity_at | latest_answer_at | score_obtained_at | validated_at | started_at |
      | 5577006791947779410 | 102      | 60      | 0              | 0           | done                     | null               | null             | null              | null         | null       |
