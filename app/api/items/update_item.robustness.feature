Feature: Update item - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'items':
      | id |
      | 4  |
      | 21 |
      | 22 |
      | 50 |
      | 60 |
    And the database has the following table 'items_items':
      | id | parent_item_id | child_item_id | child_order |
      | 1  | 4              | 21            | 0           |
      | 2  | 21             | 50            | 0           |
    And the database has the following table 'items_ancestors':
      | id | ancestor_item_id | child_item_id |
      | 1  | 4                | 21            |
      | 2  | 21               | 50            |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | manager_access | cached_manager_access | owner_access |
      | 41 | 11       | 21      | true           | true                  | false        |
      | 42 | 11       | 22      | false          | false                 | false        |
      | 43 | 11       | 4       | true           | true                  | false        |
      | 44 | 11       | 50      | true           | true                  | false        |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 71 | 11                | 11             | 1       |
    And the database has the following table 'languages':
      | id |
      | 3  |

  Scenario: default_language_id is not a number
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "default_language_id": "sewrwer3"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "default_language_id": ["decoding error: strconv.ParseInt: parsing \"sewrwer3\": invalid syntax"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: default_language_id doesn't exist
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "default_language_id": "404"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "default_language_id": ["default language should exist and there should be item's strings in this language"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: No strings in default_language_id
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "default_language_id": "3"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "default_language_id": ["default language should exist and there should be item's strings in this language"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Invalid item_id
    And I am the user with group_id "11"
    When I send a PUT request to "/items/abc" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: The user doesn't exist
    And I am the user with group_id "121"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: The user doesn't have rights to manage the item
    And I am the user with group_id "11"
    When I send a PUT request to "/items/60" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response code should be 403
    And the response error message should contain "No access rights to manage the item"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong full_screen
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "full_screen": "wrong value"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "full_screen": ["full_screen must be one of [forceYes forceNo default]"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong type
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "type": "Wrong"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "type": ["type must be one of [Root Category Chapter Task Course]"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong validation_type
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "validation_type": "Wrong"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "validation_type": ["validation_type must be one of [None All AllButOne Categories One Manual]"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong validation_min
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "validation_min": "Wrong"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "validation_min": ["expected type 'int32', got unconvertible type 'string'"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong unlocked_item_ids
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "unlocked_item_ids": "1,abc"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "unlocked_item_ids": ["all the IDs should exist and the user should have manager/owner access to them"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Non-existent id in unlocked_item_ids
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "unlocked_item_ids": "404"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "unlocked_item_ids": ["all the IDs should exist and the user should have manager/owner access to them"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: unlocked_item_ids not owned/managed by the user
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "unlocked_item_ids": "22"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "unlocked_item_ids": ["all the IDs should exist and the user should have manager/owner access to them"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong contest_entering_condition
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "contest_entering_condition": "Wrong"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "contest_entering_condition": ["contest_entering_condition must be one of [All Half One None]"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong duration (wrong format)
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "duration": "12:34"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "duration": ["invalid duration"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong duration (negative hours)
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "duration": "-1:34:56"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "duration": ["invalid duration"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong duration (too many hours)
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "duration": "839:34:56"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "duration": ["invalid duration"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong duration (negative minutes)
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "duration": "99:-1:56"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "duration": ["invalid duration"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong duration (too many minutes)
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "duration": "99:60:56"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "duration": ["invalid duration"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong duration (negative seconds)
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "duration": "99:59:-1"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "duration": ["invalid duration"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong duration (too many seconds)
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "duration": "99:59:60"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "duration": ["invalid duration"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Wrong contest_phase
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "contest_phase": "Wrong"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "contest_phase": ["contest_phase must be one of [Running Analysis Closed]"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Non-unique children item IDs
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "children": [
          {"item_id": "21", "order": 1},
          {"item_id": "21", "order": 2}
        ]
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "children": ["children IDs should be unique and the user should have manager/owner access to them"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: User doesn't have manager/owner access to children items
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "children": [
          {"item_id": "4", "order": 1},
          {"item_id": "22", "order": 2}
        ]
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid input data",
        "errors":{
          "children": ["children IDs should be unique and the user should have manager/owner access to them"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: The item is among child items
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "children": [
          {"item_id": "50", "order": 1}
        ]
      }
      """
    Then the response code should be 403
    And the response error message should contain "An item cannot become an ancestor of itself"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: The item is a descendant of a child item
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "children": [
          {"item_id": "21", "order": 1}
        ]
      }
      """
    Then the response code should be 403
    And the response error message should contain "An item cannot become an ancestor of itself"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged
