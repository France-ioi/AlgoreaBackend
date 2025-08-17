Feature: Delete an item string entry

  Background:
    Given the database has the following user:
      | group_id | login |
      | 11       | jdoe  |
    And the database has the following table "items":
      | id | default_language_tag |
      | 21 | en                   |
      | 50 | en                   |
      | 60 | sl                   |
    And the database has the following table "items_strings":
      | item_id | language_tag | title  | image_url                  | subtitle         | description        |
      | 50      | en           | Item 1 | http://myurl.com/item1.jpg | Item 1 Subtitle  | Item 1 Description |
      | 50      | sl           | Item 1 | http://myurl.com/item1.jpg | Item 1 Podnaslov | Item 1 Opis        |
      | 60      | en           | Item 2 | http://myurl.com/item2.jpg | Item 2 Subtitle  | Item 2 Description |
      | 60      | sl           | Item 2 | http://myurl.com/item2.jpg | Item 2 Podnaslov | Item 2 Opis        |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_edit_generated | is_owner_generated |
      | 11       | 21      | content            | none               | false              |
      | 11       | 50      | solution           | all                | true               |
      | 11       | 60      | content            | all                | true               |
    And the database has the following table "languages":
      | tag |
      | en  |
      | sl  |

  Scenario: Delete a language string with non-default language
    Given I am the user with id "11"
    When I send a DELETE request to "/items/50/strings/sl"
    Then the response should be "deleted"
    And the table "items_strings" should remain unchanged, except that the row with language_tag "sl" should be deleted
