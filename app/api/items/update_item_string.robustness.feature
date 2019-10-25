Feature: Update item strings - robustness

  Background:
    Given the database has the following table 'groups':
      | id | name       | type      |
      | 11 | jdoe       | UserSelf  |
      | 12 | jdoe-admin | UserAdmin |
    And the database has the following table 'users':
      | login | temp_user | group_id | owned_group_id |
      | jdoe  | 0         | 11       | 12             |
    And the database has the following table 'items':
      | id | default_language_id |
      | 50 | 2                   |
      | 60 | 3                   |
    And the database has the following table 'items_strings':
      | item_id | language_id | title  | image_url                  | subtitle        | description        |
      | 50      | 2           | Item 2 | http://myurl.com/item2.jpg | Item 2 Subtitle | Item 2 Description |
      | 50      | 3           | Item 3 | http://myurl.com/item3.jpg | Item 3 Subtitle | Item 3 Description |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | manager_access | owner_access |
      | 40 | 11       | 50      | false          | true         |
      | 41 | 11       | 21      | true           | false        |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 71 | 11                | 11             | 1       |
      | 72 | 12                | 12             | 1       |
    And the database has the following table 'languages':
      | id |
      | 2  |
      | 3  |

  Scenario: User not found
    Given I am the user with group_id "404"
    When I send a PUT request to "/items/50/strings/default" with the following body:
      """
      {
        "title": "The title"
      }
      """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "items_strings" should stay unchanged

  Scenario: Invalid item_id
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/abc/strings/default" with the following body:
      """
      {
        "title": "The title"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "items_strings" should stay unchanged

  Scenario: The title is too long
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50/strings/default" with the following body:
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
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50/strings/default" with the following body:
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
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50/strings/default" with the following body:
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
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50/strings/404" with the following body:
      """
      {
      }
      """
    Then the response code should be 400
    And the response error message should contain "No such language"
    And the table "items_strings" should stay unchanged

  Scenario: Invalid language_id
    Given I am the user with group_id "11"
    When I send a PUT request to "/items/50/strings/abc" with the following body:
      """
      {
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for language_id (should be int64)"
    And the table "items_strings" should stay unchanged

  Scenario: The user doesn't have rights to manage the item
    And I am the user with group_id "11"
    When I send a PUT request to "/items/60/strings/default" with the following body:
      """
      {
        "title": "The title"
      }
      """
    Then the response code should be 403
    And the response error message should contain "No access rights to manage the item"
