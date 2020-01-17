Feature: Update item - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 4  | fr                   |
      | 21 | fr                   |
      | 22 | fr                   |
      | 50 | fr                   |
      | 60 | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 4              | 21            | 0           |
      | 21             | 50            | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 4                | 21            |
      | 21               | 50            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_edit_generated | is_owner_generated |
      | 11       | 21      | solution           | none               | false              |
      | 11       | 22      | none               | children           | false              |
      | 11       | 4       | solution           | none               | false              |
      | 11       | 50      | solution           | all                | false              |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | can_edit | is_owner | source_group_id |
      | 11       | 4       | solution | none     | false    | 11              |
      | 11       | 21      | solution | none     | false    | 11              |
      | 11       | 50      | solution | all      | false    | 11              |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 71 | 11                | 11             | 1       |
    And the database has the following table 'languages':
      | tag |
      | sl  |

  Scenario: default_language_tag is not a string
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "default_language_tag": 1234
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
          "default_language_tag": ["expected type 'string', got unconvertible type 'float64'"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: default_language_tag doesn't exist
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "default_language_tag": "unknown"
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
          "default_language_tag": ["default language should exist and there should be item's strings in this language"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: No strings in default_language_tag
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "default_language_tag": "sl"
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
          "default_language_tag": ["default language should exist and there should be item's strings in this language"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: Invalid item_id
    And I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: The user doesn't exist
    And I am the user with id "121"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: The user doesn't have rights to edit the item
    And I am the user with id "11"
    When I send a PUT request to "/items/60" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response code should be 403
    And the response error message should contain "No access rights to edit the item"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: The user doesn't have rights to edit the item (can_edit = children)
    And I am the user with id "11"
    When I send a PUT request to "/items/22" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response code should be 403
    And the response error message should contain "No access rights to edit the item"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong full_screen
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong type
    Given I am the user with id "11"
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
          "type": ["type must be one of [Chapter Task Course]"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong validation_type
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong contest_entering_condition
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong duration (wrong format)
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong duration (negative hours)
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong duration (too many hours)
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong duration (negative minutes)
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong duration (too many minutes)
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong duration (negative seconds)
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Wrong duration (too many seconds)
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: Non-unique children item IDs
    Given I am the user with id "11"
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
          "children": ["children IDs should be unique and each should be visible to the user"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: Children items are not visible to the user
    Given I am the user with id "11"
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
          "children": ["children IDs should be unique and each should be visible to the user"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: The item is among child items
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged

  Scenario: The item is a descendant of a child item
    Given I am the user with id "11"
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
    And the table "permissions_granted" should stay unchanged
