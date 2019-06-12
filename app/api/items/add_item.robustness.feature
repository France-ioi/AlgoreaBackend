Feature: Add item - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf |
      | 1  | jdoe   | 0        | 11          |
    And the database has the following table 'items':
      | ID | bTeamsEditable | bNoScore |
      | 4  | false          | false    |
      | 21 | false          | false    |
      | 22 | false          | false    |
    And the database has the following table 'items_items':
      | ID | idItemParent | idItemChild |
      | 1  | 4            | 21          |
    And the database has the following table 'items_ancestors':
      | ID | idItemAncestor | idItemChild |
      | 1  | 4              | 21          |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | bManagerAccess | bOwnerAccess |
      | 41 | 11      | 21     | true           | false        |
      | 42 | 11      | 22     | false          | false        |
      | 43 | 11      | 4      | true           | false        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf |
      | 71 | 11              | 11           | 1       |
    And the database has the following table 'languages':
      | ID |
      | 3  |

  Scenario: Missing type
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.type": ["missing field"]
        }
      }
      """

  Scenario: Missing language_id
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter"
        },
        "string": {
          "title": "my title"
        },
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
          "language_id": ["missing field"]
        }
      }
      """

  Scenario: Missing title
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter"
        },
        "language_id": "3",
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
          "string.title": ["missing field"]
        }
      }
      """

  Scenario: Missing parent_item_id
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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

  Scenario: language_id is not a number
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Course"
        },
        "language_id": "sewrwer3",
        "string": {
          "title": "my title"
        },
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
          "language_id": ["decoding error: strconv.ParseInt: parsing \"sewrwer3\": invalid syntax"]
        }
      }
      """

  Scenario: language_id doesn't exist
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Course"
        },
        "language_id": "404",
        "string": {
          "title": "my title"
        },
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
          "language_id": ["no such language"]
        }
      }
      """

  Scenario: parent_item_id is not a number
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "id": "2",
        "item": {
          "type": "Course"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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

  Scenario: Non-existing parent
    Given I am the user with ID "1"
    And I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Course"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "parent_item_id": ["should exist and the user should have manager/owner access to it"]
        }
      }
      """

  Scenario: Not enough perm on parent
    Given the database has the following table 'groups':
      | ID | sName      | sTextId | iGrade | sType     | iVersion |
      | 11 | jdoe       |         | -2     | UserAdmin | 0        |
    And I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Course"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
        "parent_item_id": "22",
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
          "parent_item_id": ["should exist and the user should have manager/owner access to it"]
        }
      }
      """

  Scenario: The user doesn't exist
    And I am the user with ID "121"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Course"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
        "parent_item_id": "21",
        "order": 100
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong full_screen
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Course",
          "full_screen": "wrong value"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.full_screen": ["full_screen must be one of [forceYes forceNo default]"]
        }
      }
      """

  Scenario: Wrong type
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Wrong"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.type": ["type must be one of [Root Category Chapter Task Course]"]
        }
      }
      """

  Scenario: Wrong validation_type
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "validation_type": "Wrong"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.validation_type": ["validation_type must be one of [None All AllButOne Categories One Manual]"]
        }
      }
      """

  Scenario: Wrong validation_min
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "validation_min": "Wrong"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.validation_min": ["expected type 'int32', got unconvertible type 'string'"]
        }
      }
      """

  Scenario: Wrong unlocked_item_ids
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "unlocked_item_ids": "1,abc"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.unlocked_item_ids": ["all the IDs should exist and the user should have manager/owner access to them"]
        }
      }
      """

  Scenario: Non-existent ID in unlocked_item_ids
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "unlocked_item_ids": "404"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.unlocked_item_ids": ["all the IDs should exist and the user should have manager/owner access to them"]
        }
      }
      """

  Scenario: unlocked_item_ids not owned/managed by the user
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "unlocked_item_ids": "22"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.unlocked_item_ids": ["all the IDs should exist and the user should have manager/owner access to them"]
        }
      }
      """

  Scenario: Wrong team_mode
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "team_mode": "Wrong"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.team_mode": ["team_mode must be one of [All Half One None]"]
        }
      }
      """

  Scenario: Non-existent group ID in team_in_group_id
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "team_in_group_id": "404"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.team_in_group_id": ["should exist and be owned by the user"]
        }
      }
      """

  Scenario: team_in_group_id is not owned by the user
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "team_in_group_id": "11"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.team_in_group_id": ["should exist and be owned by the user"]
        }
      }
      """

  Scenario: Wrong duration (wrong format)
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "duration": "12:34"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.duration": ["invalid duration"]
        }
      }
      """

  Scenario: Wrong duration (negative hours)
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "duration": "-1:34:56"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.duration": ["invalid duration"]
        }
      }
      """

  Scenario: Wrong duration (too many hours)
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "duration": "100:34:56"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.duration": ["invalid duration"]
        }
      }
      """

  Scenario: Wrong duration (negative minutes)
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "duration": "99:-1:56"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.duration": ["invalid duration"]
        }
      }
      """

  Scenario: Wrong duration (too many minutes)
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "duration": "99:60:56"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.duration": ["invalid duration"]
        }
      }
      """

  Scenario: Wrong duration (negative seconds)
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "duration": "99:59:-1"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.duration": ["invalid duration"]
        }
      }
      """

  Scenario: Wrong duration (too many seconds)
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "duration": "99:59:60"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.duration": ["invalid duration"]
        }
      }
      """

  Scenario: Wrong contest_phase
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter",
          "contest_phase": "Wrong"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "item.contest_phase": ["contest_phase must be one of [Running Analysis Closed]"]
        }
      }
      """

  Scenario: Non-unique children item IDs
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "children": ["children IDs should be unique and the user should have manager/owner access to them"]
        }
      }
      """

  Scenario: User doesn't have manager/owner access to children items
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
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
          "children": ["children IDs should be unique and the user should have manager/owner access to them"]
        }
      }
      """

  Scenario: The parent is a child item
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
        "parent_item_id": "21",
        "order": 100,
        "children": [
          {"item_id": "21", "order": 1}
        ]
      }
      """
    Then the response code should be 403
    And the response error message should contain "An item cannot become an ancestor of itself"

  Scenario: The parent is a descendant of a child item
    Given I am the user with ID "1"
    When I send a POST request to "/items/" with the following body:
      """
      {
        "item": {
          "type": "Chapter"
        },
        "language_id": "3",
        "string": {
          "title": "my title"
        },
        "parent_item_id": "21",
        "order": 100,
        "children": [
          {"item_id": "4", "order": 1}
        ]
      }
      """
    Then the response code should be 403
    And the response error message should contain "An item cannot become an ancestor of itself"
