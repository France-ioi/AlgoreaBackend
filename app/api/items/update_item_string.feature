Feature: Update item strings

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
      | ID | idItem | idLanguage | sTitle    | sImageUrl                   | sSubtitle       | sDescription       |
      | 1  | 50     | 2          | Item 2    | http://myurl.com/item2.jpg  | Item 2 Subtitle | Item 2 Description |
      | 2  | 50     | 3          | Item 3    | http://myurl.com/item3.jpg  | Item 3 Subtitle | Item 3 Description |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | bManagerAccess | bOwnerAccess |
      | 40 | 11      | 50     | false          | true         |
      | 41 | 11      | 21     | true           | false        |
      | 42 | 11      | 60     | false          | true         |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf |
      | 71 | 11              | 11           | 1       |
      | 72 | 12              | 12           | 1       |
    And the database has the following table 'languages':
      | ID |
      | 2  |
      | 3  |

  Scenario: Update the default language string
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50/string/default" with the following body:
      """
      {
        "title": "The title",
        "image_url": "http://mysite.com/image.jpg",
        "subtitle": "The subtitle",
        "description": "The description"
      }
      """
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged but the row with ID "1"
    And the table "items_strings" at ID "1" should be:
      | ID | idItem | idLanguage | sTitle    | sImageUrl                   | sSubtitle    | sDescription    |
      | 1  | 50     | 2          | The title | http://mysite.com/image.jpg | The subtitle | The description |

  Scenario: Update the specified language string
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50/string/3" with the following body:
      """
      {
        "title": "The title",
        "image_url": "http://mysite.com/image.jpg",
        "subtitle": "The subtitle",
        "description": "The description"
      }
      """
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged but the row with ID "2"
    And the table "items_strings" at ID "2" should be:
      | ID | idItem | idLanguage | sTitle    | sImageUrl                   | sSubtitle    | sDescription    |
      | 2  | 50     | 3          | The title | http://mysite.com/image.jpg | The subtitle | The description |

  Scenario: Insert the default language string
    Given I am the user with ID "1"
    When I send a PUT request to "/items/60/string/default" with the following body:
      """
      {
        "title": "The title",
        "image_url": "http://mysite.com/image.jpg",
        "subtitle": "The subtitle",
        "description": "The description"
      }
      """
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged but the row with ID "5577006791947779410"
    And the table "items_strings" at ID "5577006791947779410" should be:
      | ID                  | idItem | idLanguage | sTitle    | sImageUrl                   | sSubtitle    | sDescription    |
      | 5577006791947779410 | 60     | 3          | The title | http://mysite.com/image.jpg | The subtitle | The description |

  Scenario: Insert the specified language string
    Given I am the user with ID "1"
    When I send a PUT request to "/items/60/string/2" with the following body:
      """
      {
        "title": "The title",
        "image_url": "http://mysite.com/image.jpg",
        "subtitle": "The subtitle",
        "description": "The description"
      }
      """
    Then the response should be "updated"
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged but the row with ID "5577006791947779410"
    And the table "items_strings" at ID "5577006791947779410" should be:
      | ID                  | idItem | idLanguage | sTitle    | sImageUrl                   | sSubtitle       | sDescription       |
      | 5577006791947779410 | 60     | 2          | The title | http://mysite.com/image.jpg | The subtitle    | The description    |

  Scenario: Valid without any fields
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50/string/default" with the following body:
      """
      {
      }
      """
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged
