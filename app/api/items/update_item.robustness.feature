Feature: Update item - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf |
      | 1  | jdoe   | 0        | 11          |
    And the database has the following table 'items':
      | ID |
      | 4  |
      | 21 |
      | 22 |
      | 50 |
      | 60 |
    And the database has the following table 'items_items':
      | ID | idItemParent | idItemChild |
      | 1  | 4            | 21          |
      | 2  | 21           | 50          |
    And the database has the following table 'items_ancestors':
      | ID | idItemAncestor | idItemChild |
      | 1  | 4              | 21          |
      | 2  | 21             | 50          |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | bManagerAccess | bCachedManagerAccess | bOwnerAccess |
      | 41 | 11      | 21     | true           | true                 | false        |
      | 42 | 11      | 22     | false          | false                | false        |
      | 43 | 11      | 4      | true           | true                 | false        |
      | 44 | 11      | 50     | true           | true                 | false        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf |
      | 71 | 11              | 11           | 1       |
    And the database has the following table 'languages':
      | ID |
      | 3  |

  Scenario: default_language_id is not a number
    Given I am the user with ID "1"
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

  Scenario: default_language_id doesn't exist
    Given I am the user with ID "1"
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

  Scenario: No strings in default_language_id
    Given I am the user with ID "1"
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

  Scenario: Invalid item_id
    And I am the user with ID "1"
    When I send a PUT request to "/items/abc" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: The user doesn't exist
    And I am the user with ID "121"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user doesn't have rights to manage the item
    And I am the user with ID "1"
    When I send a PUT request to "/items/60" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response code should be 403
    And the response error message should contain "No access rights to manage the item"

  Scenario: Wrong full_screen
    Given I am the user with ID "1"
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

  Scenario: Wrong type
    Given I am the user with ID "1"
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

  Scenario: Wrong validation_type
    Given I am the user with ID "1"
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

  Scenario: Wrong validation_min
    Given I am the user with ID "1"
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

  Scenario: Wrong unlocked_item_ids
    Given I am the user with ID "1"
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

  Scenario: Non-existent ID in unlocked_item_ids
    Given I am the user with ID "1"
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

  Scenario: unlocked_item_ids not owned/managed by the user
    Given I am the user with ID "1"
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

  Scenario: Wrong team_mode
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "team_mode": "Wrong"
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
          "team_mode": ["team_mode must be one of [All Half One None]"]
        }
      }
      """

  Scenario: Non-existent group ID in team_in_group_id
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "team_in_group_id": "404"
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
          "team_in_group_id": ["should exist and be owned by the user"]
        }
      }
      """

  Scenario: team_in_group_id is not owned by the user
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "team_in_group_id": "11"
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
          "team_in_group_id": ["should exist and be owned by the user"]
        }
      }
      """

  Scenario: Wrong duration (wrong format)
    Given I am the user with ID "1"
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

  Scenario: Wrong duration (negative hours)
    Given I am the user with ID "1"
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

  Scenario: Wrong duration (too many hours)
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "duration": "100:34:56"
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

  Scenario: Wrong duration (negative minutes)
    Given I am the user with ID "1"
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

  Scenario: Wrong duration (too many minutes)
    Given I am the user with ID "1"
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

  Scenario: Wrong duration (negative seconds)
    Given I am the user with ID "1"
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

  Scenario: Wrong duration (too many seconds)
    Given I am the user with ID "1"
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

  Scenario: Wrong contest_phase
    Given I am the user with ID "1"
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

  Scenario: Non-unique children item IDs
    Given I am the user with ID "1"
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

  Scenario: User doesn't have manager/owner access to children items
    Given I am the user with ID "1"
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

  Scenario: The item is among child items
    Given I am the user with ID "1"
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

  Scenario: The item is a descendant of a child item
    Given I am the user with ID "1"
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
