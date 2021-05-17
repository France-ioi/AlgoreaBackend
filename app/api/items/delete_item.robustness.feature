Feature: Delete an item - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    | type    | root_activity_id | root_skill_id |
      | 10 | Friends | Friends | null             | null          |
      | 11 | jdoe    | User    | null             | null          |
    And the database has the following table 'users':
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'items_propagate':
      | id | ancestors_computation_state |
      | 20 | done                        |
      | 21 | done                        |
      | 22 | done                        |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 20 | fr                   |
      | 21 | fr                   |
      | 22 | fr                   |
    And the database has the following table 'permissions_propagate':
      | group_id | item_id |
      | 10       | 21      |
      | 10       | 22      |
      | 11       | 21      |
      | 11       | 22      |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | is_owner_generated |
      | 10       | 21      | true               |
      | 11       | 22      | false              |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | is_owner | source_group_id |
      | 10       | 21      | true     | 10              |
      | 11       | 22      | false    | 10              |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 10              | 11             |
    And the groups ancestors are computed
    And the database has the following table 'attempts':
      | id | participant_id | root_item_id |
      | 0  | 10             | null         |
      | 1  | 10             | 22           |
    And the database has the following table 'answers':
      | participant_id | attempt_id | item_id | author_id | created_at          |
      | 10             | 0          | 21      | 10        | 2019-05-30 11:00:00 |
      | 10             | 0          | 22      | 10        | 2019-05-30 11:00:00 |
      | 10             | 1          | 21      | 10        | 2019-05-30 11:00:00 |
    And the database has the following table 'filters':
      | id | user_id | item_id |
      | 1  | 10      | 21      |
      | 2  | 10      | 22      |
      | 3  | 11      | null    |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id |
      | 10       | 21      |
      | 11       | 22      |
    And the database has the following table 'item_dependencies':
      | item_id | dependent_item_id |
      | 21      | 21                |
      | 21      | 22                |
      | 22      | 21                |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 20               | 21            |
      | 20               | 22            |
      | 21               | 22            |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 20             | 21            | 1           |
      | 21             | 22            | 1           |
    And the database has the following table 'items_strings':
      | item_id | language_tag |
      | 20      | fr           |
      | 21      | fr           |
      | 22      | en           |
      | 22      | fr           |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id |
      | 0          | 10             | 21      |
      | 0          | 10             | 22      |
      | 1          | 10             | 21      |
    And the database has the following table 'threads':
      | id | creator_id | type | item_id |
      | 1  | 10         | Help | 21      |
      | 2  | 10         | Help | 22      |
      | 3  | 11         | Bug  | null    |
    And the database has the following table 'languages':
      | tag |
      | fr  |
      | en  |
      | sl  |

  Scenario: Invalid item_id
    Given I am the user with id "11"
    When I send a DELETE request to "/items/123456789012345678901234567890"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_propagate" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged
    And the table "item_dependencies" should stay unchanged
    And the table "groups_contest_items" should stay unchanged
    And the table "attempts" should stay unchanged
    And the table "results" should stay unchanged
    And the table "answers" should stay unchanged
    And the table "filters" should stay unchanged
    And the table "threads" should stay unchanged

  Scenario: The user is not an owner of the item
    Given I am the user with id "11"
    When I send a DELETE request to "/items/22"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_propagate" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged
    And the table "item_dependencies" should stay unchanged
    And the table "groups_contest_items" should stay unchanged
    And the table "attempts" should stay unchanged
    And the table "results" should stay unchanged
    And the table "answers" should stay unchanged
    And the table "filters" should stay unchanged
    And the table "threads" should stay unchanged

  Scenario: The item has children
    Given I am the user with id "11"
    When I send a DELETE request to "/items/21"
    Then the response code should be 422
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Unprocessable Entity",
      "error_text": "The item must not have children"
    }
    """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_propagate" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged
    And the table "item_dependencies" should stay unchanged
    And the table "groups_contest_items" should stay unchanged
    And the table "attempts" should stay unchanged
    And the table "results" should stay unchanged
    And the table "answers" should stay unchanged
    And the table "filters" should stay unchanged
    And the table "threads" should stay unchanged
