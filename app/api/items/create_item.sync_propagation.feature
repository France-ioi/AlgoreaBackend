Feature: Create item
  Background:
    Given the database has the following table "groups":
      | id | name    | type    | root_activity_id | root_skill_id |
      | 10 | Friends | Friends | null             | null          |
    And the database has the following user:
      | group_id | login |
      | 11       | jdoe  |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage            |
      | 10       | 11         | memberships_and_group |
    And the database has the following table "items":
      | id | default_language_tag |
      | 21 | fr                   |
      | 22 | fr                   |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order |
      | 21             | 22            | 0           |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_edit_generated |
      | 10       | 21      | content            | none               |
      | 11       | 21      | solution           | children           |
    And the database has the following table "permissions_granted":
      | group_id | item_id | can_view | can_edit | source_group_id | latest_update_at    |
      | 11       | 21      | solution | children | 11              | 2019-05-30 11:00:00 |
    And the database has the following table "attempts":
      | id | participant_id |
      | 0  | 10             |
      | 1  | 11             |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | score_computed |
      | 0          | 10             | 21      | 100            |
      | 0          | 10             | 22      | 100            |
      | 1          | 11             | 21      | 100            |
      | 1          | 11             | 22      | 100            |
    And the database has the following table "languages":
      | tag |
      | sl  |
    And the generated permissions are computed
    And the application config is:
    """
    server:
      propagation_endpoint: "/not_found" # set a non-empty propagation_endpoint pointing to nowhere to disable the async propagation
    """

  Scenario: Synchronously recomputes results of the current user linked to the parent item
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Task",
        "language_tag": "sl",
        "title": "my title 🐱",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task 🐱",
        "description": "the goal of this task is ... 🐱",
        "parent": {"item_id": "21"}
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": { "id": "5577006791947779410" }
      }
      """
    And the table "results" should remain unchanged, regardless of the rows with item_id "21"
    And the table "results" at item_id "21" should be:
      | attempt_id | participant_id | score_computed |
      | 0          | 10             | 100            |
      | 1          | 11             | 50             |
    And the table "results_propagate_sync" should be empty
    # inserted by after_insert_items_items
    And the table "results_propagate" should be:
      | participant_id | attempt_id | item_id | state            |
      | 10             | 0          | 21      | to_be_recomputed |
      | 11             | 1          | 21      | to_be_recomputed |

  Scenario: Synchronously propagates newly created permissions for the current user on the new item (no parent, no children)
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Task",
        "language_tag": "sl",
        "title": "my title",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task",
        "description": "the goal of this task is ...",
        "as_root_of_group_id": "10"
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": { "id": "5577006791947779410" }
      }
      """
    And the table "permissions_generated" should remain unchanged, regardless of the rows with group_id "11"
    And the table "permissions_generated" at group_id "11" should be:
      | item_id             | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21                  | solution           | none                     | none                | children           | 0                  |
      | 22                  | none               | none                     | none                | none               | 0                  |
      | 5577006791947779410 | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | 1                  |
    And the table "permissions_propagate" should be empty
    And the table "permissions_propagate_sync" should be empty

  Scenario: Synchronously propagates results of the current user linked to child items
    Given I am the user with id "11"
    And the database table "items" also has the following rows:
      | id | default_language_tag |
      | 12 | fr                   |
      | 34 | fr                   |
    And the database table "permissions_granted" also has the following rows:
      | group_id | item_id | can_view                 | can_grant_view      | can_watch         | can_edit       | is_owner | source_group_id | latest_update_at    |
      | 11       | 12      | content_with_descendants | solution            | answer            | all            | 0        | 11              | 2019-05-30 11:00:00 |
      | 11       | 34      | solution                 | solution_with_grant | answer_with_grant | all_with_grant | 0        | 11              | 2019-05-30 11:00:00 |
    And the generated permissions are computed
    And the database table "results" also has the following rows:
      | attempt_id | participant_id | item_id | score_computed |
      | 0          | 10             | 12      | 60             |
      | 1          | 11             | 12      | 60             |
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "language_tag": "sl",
        "title": "my title",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task",
        "description": "the goal of this task is ...",
        "as_root_of_group_id": "10",
        "children": [
          {"item_id": "12", "order": 0},
          {"item_id": "34", "order": 1, "category": "Application", "score_weight": 2}
        ]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": { "id": "5577006791947779410" }
      }
      """
    And the table "results" should remain unchanged, regardless of the rows with participant_id "11"
    And the table "results" at participant_id "11" should be:
      | attempt_id | item_id             | score_computed |
      | 1          | 12                  | 60             |
      | 1          | 21                  | 100            |
      | 1          | 22                  | 100            |
      | 1          | 5577006791947779410 | 20             |
    # inserted by after_insert_items_items
    And the table "results_propagate" should be:
      | participant_id | attempt_id | item_id | state            |
      | 10             | 0          | 12      | to_be_propagated |
      | 11             | 1          | 12      | to_be_propagated |
    And the table "results_propagate_sync" should be empty
