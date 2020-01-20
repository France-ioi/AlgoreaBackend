Feature: Add item - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'items':
      | id | teams_editable | no_score | default_language_tag |
      | 4  | false          | false    | fr                   |
      | 21 | false          | false    | fr                   |
      | 22 | false          | false    | fr                   |
      | 23 | false          | false    | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 4              | 21            | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 4                | 21            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_edit_generated |
      | 11       | 4       | solution           | children           |
      | 11       | 21      | solution           | children           |
      | 11       | 22      | none               | none               |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | source_group_id | can_edit |
      | 11       | 4       | solution | 11              | children |
      | 11       | 21      | solution | 11              | children |
      | 11       | 23      | solution | 11              | none     |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 71 | 11                | 11             | 1       |
    And the database has the following table 'languages':
      | tag |
      | sl  |

  Scenario: Missing type
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
          "type": ["missing field"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Missing language_tag
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
          "language_tag": ["missing field"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Missing title
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "language_tag": "sl",
        "parent_item_id": "21",
        "order": 100
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
          "title": ["missing field"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Missing parent_item_id
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "language_tag": "sl",
        "title": "my title",
        "order": 100
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
          "parent_item_id": ["missing field"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: language_tag is not a string
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": 123,
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
          "language_tag": ["expected type 'string', got unconvertible type 'float64'"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: language_tag doesn't exist
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": "unknown",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
          "language_tag": ["no such language"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: parent_item_id is not a number
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "id": "2",
        "type": "Course",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "sfaewr20",
        "order": 100
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
          "parent_item_id": ["decoding error: strconv.ParseInt: parsing \"sfaewr20\": invalid syntax"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Non-existing parent
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "404",
        "order": 100
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
          "parent_item_id": ["should exist and the user should be able to manage its children"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario Outline: Not enough perm on parent
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "<parent_item>",
        "order": 100
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
          "parent_item_id": ["should exist and the user should be able to manage its children"]
        }
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged
  Examples:
    | parent_item |
    | 22          |
    | 23          |

  Scenario: The user doesn't exist
    And I am the user with id "121"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
      }
      """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong full_screen
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "full_screen": "wrong value",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong type
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Wrong",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong validation_type
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "validation_type": "Wrong",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong contest_entering_condition
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "contest_entering_condition": "Wrong",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong duration (wrong format)
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "duration": "12:34",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong duration (negative hours)
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "duration": "-1:34:56",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong duration (too many hours)
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "duration": "839:34:56",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong duration (negative minutes)
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "duration": "99:-1:56",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong duration (too many minutes)
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "duration": "99:60:56",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong duration (negative seconds)
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "duration": "99:59:-1",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Wrong duration (too many seconds)
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "duration": "99:59:60",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Non-unique children item IDs
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "4",
        "order": 100,
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Children items are not visible to the user
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100,
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
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: The parent is a child item
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100,
        "children": [
          {"item_id": "21", "order": 1}
        ]
      }
      """
    Then the response code should be 403
    And the response error message should contain "An item cannot become an ancestor of itself"
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: The parent is a descendant of a child item
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "order": 100,
        "children": [
          {"item_id": "4", "order": 1}
        ]
      }
      """
    Then the response code should be 403
    And the response error message should contain "An item cannot become an ancestor of itself"
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged
