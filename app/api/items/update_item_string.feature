Feature: Update item strings

  Background:
    Given the database has the following table 'groups':
      | id | name | type     |
      | 11 | jdoe | UserSelf |
    And the database has the following table 'users':
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'items':
      | id | default_language_id |
      | 21 | 2                   |
      | 50 | 2                   |
      | 60 | 3                   |
    And the database has the following table 'items_strings':
      | id | item_id | language_id | title  | image_url                  | subtitle        | description        |
      | 1  | 50      | 2           | Item 2 | http://myurl.com/item2.jpg | Item 2 Subtitle | Item 2 Description |
      | 2  | 50      | 3           | Item 3 | http://myurl.com/item3.jpg | Item 3 Subtitle | Item 3 Description |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_edit_generated | is_owner_generated |
      | 11       | 21      | solution           | none               | false              |
      | 11       | 50      | solution           | all                | true               |
      | 11       | 60      | solution           | all                | true               |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 71 | 11                | 11             | 1       |
    And the database has the following table 'languages':
      | id |
      | 2  |
      | 3  |

  Scenario: Update the default language string
    Given I am the user with id "11"
    When I send a PUT request to "/items/50/strings/default" with the following body:
      """
      {
        "title": "The title",
        "image_url": "http://mysite.com/image.jpg",
        "subtitle": "The subtitle",
        "description": "The description"
      }
      """
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged but the row with id "1"
    And the table "items_strings" at id "1" should be:
      | id | item_id | language_id | title     | image_url                   | subtitle     | description     |
      | 1  | 50      | 2           | The title | http://mysite.com/image.jpg | The subtitle | The description |

  Scenario: Update the specified language string
    Given I am the user with id "11"
    When I send a PUT request to "/items/50/strings/3" with the following body:
      """
      {
        "title": "The title",
        "image_url": "http://mysite.com/image.jpg",
        "subtitle": "The subtitle",
        "description": "The description"
      }
      """
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged but the row with id "2"
    And the table "items_strings" at id "2" should be:
      | id | item_id | language_id | title     | image_url                   | subtitle     | description     |
      | 2  | 50      | 3           | The title | http://mysite.com/image.jpg | The subtitle | The description |

  Scenario: Insert the default language string
    Given I am the user with id "11"
    When I send a PUT request to "/items/60/strings/default" with the following body:
      """
      {
        "title": "The title",
        "image_url": "http://mysite.com/image.jpg",
        "subtitle": "The subtitle",
        "description": "The description"
      }
      """
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged but the row with id "5577006791947779410"
    And the table "items_strings" at id "5577006791947779410" should be:
      | id                  | item_id | language_id | title     | image_url                   | subtitle     | description     |
      | 5577006791947779410 | 60      | 3           | The title | http://mysite.com/image.jpg | The subtitle | The description |

  Scenario: Insert the specified language string
    Given I am the user with id "11"
    When I send a PUT request to "/items/60/strings/2" with the following body:
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
    And the table "items_strings" should stay unchanged but the row with id "5577006791947779410"
    And the table "items_strings" at id "5577006791947779410" should be:
      | id                  | item_id | language_id | title     | image_url                   | subtitle     | description     |
      | 5577006791947779410 | 60      | 2           | The title | http://mysite.com/image.jpg | The subtitle | The description |

  Scenario: Valid without any fields
    Given I am the user with id "11"
    When I send a PUT request to "/items/50/strings/default" with the following body:
      """
      {
      }
      """
    Then the response should be "updated"
    And the table "items_strings" should stay unchanged
