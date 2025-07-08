Feature: Search for items
  Background:
    Given the database has the following table "items":
      | id | type    | default_language_tag |
      | 1  | Chapter | en                   |
      | 2  | Task    | en                   |
      | 3  | Task    | en                   |
      | 4  | Skill   | en                   |
      | 6  | Chapter | en                   |
      | 7  | Task    | en                   |
      | 10 | Chapter | en                   |
      | 11 | Chapter | en                   |
      | 12 | Chapter | en                   |
      | 13 | Chapter | en                   |
      | 14 | Chapter | en                   |
      | 15 | Chapter | en                   |
      | 16 | Chapter | en                   |
      | 17 | Chapter | en                   |
      | 18 | Chapter | en                   |
      | 19 | Chapter | en                   |
      | 20 | Chapter | en                   |
      | 21 | Chapter | en                   |
      | 22 | Chapter | en                   |
      | 23 | Chapter | en                   |
      | 24 | Chapter | en                   |
      | 25 | Chapter | en                   |
      | 26 | Chapter | en                   |
      | 27 | Chapter | en                   |
      | 28 | Chapter | en                   |
      | 29 | Chapter | en                   |
      | 30 | Chapter | en                   |
      | 31 | Chapter | en                   |
    And the database has the following table "items_strings":
      | item_id | language_tag | title                                     |
      | 1       | fr           | amazing Chapter                           |
      | 2       | fr           | amazing Task                              |
      | 3       | en           | amazing Task                              |
      | 4       | fr           | amazing \|\|\|Our Skill \\\\\\%\\\\%\\ :) |
      | 6       | en           | Another amazing Chapter                   |
      | 6       | fr           | Un autre chapitre                         |
      | 7       | en           | Another amazing Task                      |
      | 10      | en           | amazing third chapter                     |
      | 10      | fr           | Le troisième chapitre                     |
      | 11      | en           | chapter                                   |
      | 12      | en           | chapter                                   |
      | 13      | en           | chapter                                   |
      | 14      | en           | chapter                                   |
      | 15      | en           | chapter                                   |
      | 16      | en           | chapter                                   |
      | 17      | en           | chapter                                   |
      | 18      | en           | chapter                                   |
      | 19      | en           | chapter                                   |
      | 20      | en           | chapter                                   |
      | 21      | en           | chapter                                   |
      | 22      | en           | chapter                                   |
      | 23      | en           | chapter                                   |
      | 24      | en           | chapter                                   |
      | 25      | en           | chapter                                   |
      | 26      | en           | chapter                                   |
      | 27      | en           | chapter                                   |
      | 28      | en           | chapter                                   |
      | 29      | en           | chapter                                   |
      | 30      | en           | chapter                                   |
      | 31      | en           | chapter                                   |
    And the database has the following users:
      | login | default_language | temp_user | group_id | first_name  | last_name | grade |
      | owner | fr               | 0         | 21       | Jean-Michel | Blanquer  | 3     |
    And the groups ancestors are computed
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       |
      | 22       | 1       | info                     |
      | 21       | 2       | content                  |
      | 21       | 3       | none                     |
      | 21       | 4       | info                     |
      | 21       | 6       | solution                 |
      | 21       | 7       | info                     |
      | 21       | 10      | content_with_descendants |
      | 21       | 11      | content_with_descendants |
      | 21       | 12      | content_with_descendants |
      | 21       | 13      | content_with_descendants |
      | 21       | 14      | content_with_descendants |
      | 21       | 15      | content_with_descendants |
      | 21       | 16      | content_with_descendants |
      | 21       | 17      | content_with_descendants |
      | 21       | 18      | content_with_descendants |
      | 21       | 19      | content_with_descendants |
      | 21       | 20      | content_with_descendants |
      | 21       | 21      | content_with_descendants |
      | 21       | 22      | content_with_descendants |
      | 21       | 23      | content_with_descendants |
      | 21       | 24      | content_with_descendants |
      | 21       | 25      | content_with_descendants |
      | 21       | 26      | content_with_descendants |
      | 21       | 27      | content_with_descendants |
      | 21       | 28      | content_with_descendants |
      | 21       | 29      | content_with_descendants |
      | 21       | 30      | content_with_descendants |
      | 21       | 31      | content_with_descendants |

  Scenario: Search for items with "amazing"
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=amazing"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "title": "amazing Task",
        "type": "Task",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false}
      },
      {
        "id": "4",
        "title": "amazing |||Our Skill \\\\\\%\\\\%\\ :)",
        "type": "Skill",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false}
      },
      {
        "id": "7",
        "title": "Another amazing Task",
        "type": "Task",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false}
      }
    ]
    """

  Scenario: Should treat the words in the search string as "AND", and work with accents
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=chapitre%20troisième"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "10",
        "title": "Le troisième chapitre",
        "type": "Chapter",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}
      }
    ]
    """

  Scenario: Should treat the words in the search string as "AND", and work without accents
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=chapitre%20troisieme"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "10",
        "title": "Le troisième chapitre",
        "type": "Chapter",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}
      }
    ]
    """

  Scenario: Search for items with "amazing" (limit=2)
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=amazing&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "title": "amazing Task",
        "type": "Task",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false}
      },
      {
        "id": "4",
        "title": "amazing |||Our Skill \\\\\\%\\\\%\\ :)",
        "type": "Skill",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false}
      }
    ]
    """

  Scenario: Search for items with "amazing", include only items of specified types
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=amazing&types_include=Task"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "title": "amazing Task",
        "type": "Task",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false}
      },
      {
        "id": "7",
        "title": "Another amazing Task",
        "type": "Task",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false}
      }
    ]
    """

  Scenario: Search for items with "amazing", exclude items of specified types
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=amazing&types_exclude=Task"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "title": "amazing |||Our Skill \\\\\\%\\\\%\\ :)",
        "type": "Skill",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false}
      }
    ]
    """

  Scenario: Search for items with "amazing", combination of types_include & types_exclude
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=amazing&types_include=Task,Skill&types_exclude=Task"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "title": "amazing |||Our Skill \\\\\\%\\\\%\\ :)",
        "type": "Skill",
        "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false}
      }
    ]
    """

  Scenario: Check the default limit
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=chapter"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "11", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "12", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "13", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "14", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "15", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "16", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "17", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "18", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "19", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "20", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "21", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "22", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "23", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "24", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "25", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "26", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "27", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "28", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "29", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}},
      {"id": "30", "title": "chapter", "type": "Chapter",
       "permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false}}
    ]
    """
