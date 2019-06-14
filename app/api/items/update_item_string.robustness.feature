Feature: Update item strings - robustness

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned |
      | 1  | jdoe   | 0        | 11          | 12           |
    And the database has the following table 'groups':
      | ID | sName      | sType     |
      | 11 | jdoe       | UserSelf  |
      | 12 | jdoe-admin | UserAdmin |
    And the database has the following table 'items':
      | ID | idDefaultLanguage |
      | 50 | 2                 |
      | 60 | 3                 |
    And the database has the following table 'items_strings':
      | idItem | idLanguage | sTitle | sImageUrl                  | sSubtitle       | sDescription       |
      | 50     | 2          | Item 2 | http://myurl.com/item2.jpg | Item 2 Subtitle | Item 2 Description |
      | 50     | 3          | Item 3 | http://myurl.com/item3.jpg | Item 3 Subtitle | Item 3 Description |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | bManagerAccess | bOwnerAccess |
      | 40 | 11      | 50     | false          | true         |
      | 41 | 11      | 21     | true           | false        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf |
      | 71 | 11              | 11           | 1       |
      | 72 | 12              | 12           | 1       |
    And the database has the following table 'languages':
      | ID |
      | 2  |
      | 3  |

  Scenario: User not found
    Given I am the user with ID "404"
    When I send a PUT request to "/items/50/string/default" with the following body:
      """
      {
        "title": "The title"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "items_strings" should stay unchanged

  Scenario: Invalid item_id
    Given I am the user with ID "1"
    When I send a PUT request to "/items/abc/string/default" with the following body:
      """
      {
        "title": "The title"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "items_strings" should stay unchanged

  Scenario: The title is too long
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50/string/default" with the following body:
      """
      {
        "title": "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901"
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
          "title": ["title must be a maximum of 200 characters in length"]
        }
      }
      """
    And the table "items_strings" should stay unchanged

  Scenario: Image URL is too long
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50/string/default" with the following body:
      """
      {
        "image_url": "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901"
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
          "image_url": ["image_url must be a maximum of 100 characters in length"]
        }
      }
      """
    And the table "items_strings" should stay unchanged

  Scenario: The subtitle is too long
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50/string/default" with the following body:
      """
      {
        "subtitle": "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901"
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
          "subtitle": ["subtitle must be a maximum of 200 characters in length"]
        }
      }
      """
    And the table "items_strings" should stay unchanged

  Scenario: Wrong language
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50/string/404" with the following body:
      """
      {
      }
      """
    Then the response code should be 400
    And the response error message should contain "No such language"
    And the table "items_strings" should stay unchanged

  Scenario: Invalid language_id
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50/string/abc" with the following body:
      """
      {
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for language_id (should be int64)"
    And the table "items_strings" should stay unchanged

  Scenario: The user doesn't have rights to manage the item
    And I am the user with ID "1"
    When I send a PUT request to "/items/60/string/default" with the following body:
      """
      {
        "title": "The title"
      }
      """
    Then the response code should be 403
    And the response error message should contain "No access rights to manage the item"
