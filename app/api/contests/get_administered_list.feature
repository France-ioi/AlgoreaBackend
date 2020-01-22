Feature: Get the contests that the user has administration rights on (contestAdminList)
  Background:
    Given the database has the following users:
      | login          | group_id | default_language |
      | possesseur     | 21       | fr               |
      | owner          | 31       | en               |
      | administrateur | 41       | fr               |
      | admin          | 51       | en               |
      | guest          | 61       | en               |
      | panas          | 71       | uk               |
    And the database has the following table 'languages':
      | tag |
      | en  |
      | fr  |
      | sl  |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 41                | 41             | 1       |
      | 42                | 42             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 61                | 61             | 1       |
      | 62                | 62             | 1       |
      | 71                | 71             | 1       |
      | 72                | 72             | 1       |
    And the database has the following table 'items':
      | id | duration | default_language_tag | allows_multiple_attempts |
      | 50 | 00:00:00 | fr                   | 0                        |
      | 60 | 00:00:01 | en                   | 1                        |
      | 10 | 00:00:02 | en                   | 0                        |
      | 70 | 00:00:03 | fr                   | 0                        |
      | 80 | 00:00:03 | sl                   | 0                        |
      | 90 | 00:00:03 | sl                   | 0                        |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 0           |
      | 10             | 70            | 1           |
      | 60             | 70            | 0           |
      | 90             | 80            | 0           |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title      |
      | 10      | en           | Chapter    |
      | 10      | fr           | Chapitre   |
      | 60      | en           | Contest    |
      | 70      | en           | Contest 2  |
      | 70      | fr           | Concours 2 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 50      | solution                 |
      | 21       | 60      | content_with_descendants |
      | 21       | 70      | content_with_descendants |
      | 31       | 50      | solution                 |
      | 31       | 60      | content_with_descendants |
      | 31       | 70      | content_with_descendants |
      | 41       | 10      | content                  |
      | 41       | 50      | solution                 |
      | 41       | 60      | solution                 |
      | 41       | 70      | content_with_descendants |
      | 51       | 10      | info                     |
      | 51       | 50      | solution                 |
      | 51       | 60      | solution                 |
      | 51       | 70      | content_with_descendants |
      | 71       | 80      | content_with_descendants |
      | 71       | 90      | content_with_descendants |

  Scenario: User's default language is French (most parents are invisible)
    Given I am the user with id "21"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "parents": [], "title": null, "language_tag": null},
      {"id": "70", "allows_multiple_attempts": false, "parents": [{"title": "Contest", "language_tag": "en"}],
       "title": "Concours 2", "language_tag": "fr"},
      {"id": "60", "allows_multiple_attempts": true, "parents": [], "title": "Contest", "language_tag": "en"}
    ]
    """

  Scenario: User's default language is English  (most parents are invisible)
    Given I am the user with id "31"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "parents": [], "title": null, "language_tag": null},
      {"id": "60", "allows_multiple_attempts": true, "parents": [], "title": "Contest", "language_tag": "en"},
      {"id": "70", "allows_multiple_attempts": false, "parents": [{"title": "Contest", "language_tag": "en"}],
       "title": "Contest 2", "language_tag": "en"}
    ]
    """

  Scenario: User's default language is French (parents are visible)
    Given I am the user with id "41"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "parents": [], "title": null, "language_tag": null},
      {"id": "70", "allows_multiple_attempts": false,
       "parents": [{"title": "Chapitre", "language_tag": "fr"}, {"title": "Contest", "language_tag": "en"}],
       "title": "Concours 2", "language_tag": "fr"},
      {"id": "60", "allows_multiple_attempts": true, "parents": [{"title": "Chapitre", "language_tag": "fr"}],
       "title": "Contest", "language_tag": "en"}
    ]
    """

  Scenario: User's default language is English  (parents are visible)
    Given I am the user with id "51"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "parents": [], "title": null, "language_tag": null},
      {"id": "60", "allows_multiple_attempts": true, "parents": [{"title": "Chapter", "language_tag": "en"}],
       "title": "Contest", "language_tag": "en"},
      {"id": "70", "allows_multiple_attempts": false,
       "parents": [{"title": "Chapter", "language_tag": "en"}, {"title": "Contest", "language_tag": "en"}],
       "title": "Contest 2", "language_tag": "en"}
    ]
    """

  Scenario: Empty result
    Given I am the user with id "61"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: User's default language is English  (parents are visible), limit=1
    Given I am the user with id "51"
    When I send a GET request to "/contests/administered?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "parents": [], "title": null, "language_tag": null}
    ]
    """

  Scenario: User's default language is English  (parents are visible), start from the second row, limit=1
    Given I am the user with id "51"
    When I send a GET request to "/contests/administered?from.title&from.id=50&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "60", "allows_multiple_attempts": true, "parents": [{"title": "Chapter", "language_tag": "en"}],
       "title": "Contest", "language_tag": "en"}
    ]
    """

  Scenario: User's default language is English  (parents are visible), inverse order
    Given I am the user with id "51"
    When I send a GET request to "/contests/administered?sort=-title,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "70", "allows_multiple_attempts": false,
       "parents": [{"title": "Chapter", "language_tag": "en"}, {"title": "Contest", "language_tag": "en"}],
       "title": "Contest 2", "language_tag": "en"},
      {"id": "60", "allows_multiple_attempts": true, "parents": [{"title": "Chapter", "language_tag": "en"}],
       "title": "Contest", "language_tag": "en"},
      {"id": "50", "allows_multiple_attempts": false, "parents": [], "title": null, "language_tag": null}
    ]
    """

  Scenario: Keeps parents with nil titles
    Given I am the user with id "71"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "80", "allows_multiple_attempts": false, "parents": [{"language_tag": null, "title": null}],
       "title": null, "language_tag": null},
      {"id": "90", "allows_multiple_attempts": false, "parents": [], "title": null, "language_tag": null}
    ]
    """
