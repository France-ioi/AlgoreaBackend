Feature: Add item - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'groups':
      | id | name    | type    |
      | 30 | Friends | Friends |
    And the database has the following table 'items':
      | id | teams_editable | no_score | default_language_tag |
      | 4  | false          | false    | fr                   |
      | 21 | false          | false    | fr                   |
      | 22 | false          | false    | fr                   |
      | 23 | false          | false    | fr                   |
      | 24 | false          | false    | fr                   |
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
      | 11       | 24      | info               | none               |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | source_group_id | can_edit |
      | 11       | 4       | solution | 11              | children |
      | 11       | 21      | solution | 11              | children |
      | 11       | 23      | solution | 11              | none     |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 30                | 30             |
      | 30                | 11             |
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

  Scenario Outline: Wrong optional field value
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "<field>": <value>
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
          "<field>": ["<error>"]
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
      | field                      | value         | error                                                                          |
      | full_screen                | "wrong value" | full_screen must be one of [forceYes forceNo default]                          |
      | type                       | "Wrong"       | type must be one of [Chapter Task Course]                                      |
      | validation_type            | "Wrong"       | validation_type must be one of [None All AllButOne Categories One Manual]      |
      | contest_entering_condition | "Wrong"       | contest_entering_condition must be one of [All Half One None]                  |
      | duration                   | "12:34"       | invalid duration                                                               |
      | duration                   | "-1:34:56"    | invalid duration                                                               |
      | duration                   | "839:34:56"   | invalid duration                                                               |
      | duration                   | "99:-1:56"    | invalid duration                                                               |
      | duration                   | "99:60:56"    | invalid duration                                                               |
      | duration                   | "99:59:-1"    | invalid duration                                                               |
      | duration                   | "99:59:60"    | invalid duration                                                               |
      | category                   | "wrong"       | category must be one of [Undefined Discovery Application Validation Challenge] |
      | score_weight               | "wrong"       | expected type 'int8', got unconvertible type 'string'                          |

  Scenario Outline: Wrong optional field value in the array of children
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "children": [{
          "item_id": 24,
          "<field>": <value>
        }]
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
          "children[0].<field>": ["<error>"]
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
      | field        | value   | error                                                                          |
      | category     | "wrong" | category must be one of [Undefined Discovery Application Validation Challenge] |
      | score_weight | "wrong" | expected type 'int8', got unconvertible type 'string'                          |

  Scenario Outline: Not enough permissions for setting propagation in items_items
    Given I am the user with id "11"
    And the database table 'items' has also the following row:
      | id | default_language_tag |
      | 90 | fr                   |
    And the database table 'permissions_generated' has also the following row:
      | group_id | item_id | <permission_column> | can_view_generated |
      | 11       | 90      | <permission_value>  | info               |
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": "sl",
        "title": "my title",
        "parent_item_id": "21",
        "children": [{
          "item_id": 90,
          "order": 1,
          "<field>": {{"<value>" != "true" && "<value>" != "false" ? "\"<value>\"" : <value>}}
        }]
      }
      """
    Then the response code should be 403
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Forbidden",
        "error_text": "<error>"
      }
      """
    And the table "items" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged
    Examples:
      | field                         | value                       | permission_column        | permission_value         | error                                                            |
      | content_view_propagation      | as_content                  | can_grant_view_generated | none                     | Not enough permissions for setting content_view_propagation      |
      | content_view_propagation      | as_info                     | can_grant_view_generated | none                     | Not enough permissions for setting content_view_propagation      |
      | upper_view_levels_propagation | as_is                       | can_grant_view_generated | content_with_descendants | Not enough permissions for setting upper_view_levels_propagation |
      | upper_view_levels_propagation | as_content_with_descendants | can_grant_view_generated | content                  | Not enough permissions for setting upper_view_levels_propagation |
      | grant_view_propagation        | true                        | can_grant_view_generated | solution                 | Not enough permissions for setting grant_view_propagation        |
      | watch_propagation             | true                        | can_watch_generated      | answer                   | Not enough permissions for setting watch_propagation             |
      | edit_propagation              | true                        | can_edit_generated       | all                      | Not enough permissions for setting edit_propagation             |

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
