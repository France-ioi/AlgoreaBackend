Feature: Get the contests that the user has administration rights on (contestAdminList)
  Background:
    Given the database has the following users:
      | login          | group_id | owned_group_id | default_language |
      | possesseur     | 21       | 22             | fr               |
      | owner          | 31       | 32             | en               |
      | administrateur | 41       | 42             | fr               |
      | admin          | 51       | 52             | en               |
      | guest          | 61       | 62             | en               |
      | panas          | 71       | 72             | uk               |
    And the database has the following table 'languages':
      | id | code |
      | 1  | en   |
      | 2  | fr   |
      | 3  | uk   |
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
      | id | duration | default_language_id | has_attempts |
      | 50 | 00:00:00 | 2                   | 0            |
      | 60 | 00:00:01 | 1                   | 1            |
      | 10 | 00:00:02 | 1                   | 0            |
      | 70 | 00:00:03 | 2                   | 0            |
      | 80 | 00:00:03 | 3                   | 0            |
      | 90 | 00:00:03 | 3                   | 0            |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 0           |
      | 10             | 70            | 1           |
      | 60             | 70            | 0           |
      | 90             | 80            | 0           |
    And the database has the following table 'items_strings':
      | item_id | language_id | title      |
      | 10      | 1           | Chapter    |
      | 10      | 2           | Chapitre   |
      | 60      | 1           | Contest    |
      | 70      | 1           | Contest 2  |
      | 70      | 2           | Concours 2 |
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
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null, "language_id": null},
      {"id": "70", "team_only_contest": false, "parents": [{"title": "Contest", "language_id": "1"}],
       "title": "Concours 2", "language_id": "2"},
      {"id": "60", "team_only_contest": true, "parents": [], "title": "Contest", "language_id": "1"}
    ]
    """

  Scenario: User's default language is English  (most parents are invisible)
    Given I am the user with group_id "31"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null, "language_id": null},
      {"id": "60", "team_only_contest": true, "parents": [], "title": "Contest", "language_id": "1"},
      {"id": "70", "team_only_contest": false, "parents": [{"title": "Contest", "language_id": "1"}],
       "title": "Contest 2", "language_id": "1"}
    ]
    """

  Scenario: User's default language is French (parents are visible)
    Given I am the user with group_id "41"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null, "language_id": null},
      {"id": "70", "team_only_contest": false,
       "parents": [{"title": "Chapitre", "language_id": "2"}, {"title": "Contest", "language_id": "1"}],
       "title": "Concours 2", "language_id": "2"},
      {"id": "60", "team_only_contest": true, "parents": [{"title": "Chapitre", "language_id": "2"}],
       "title": "Contest", "language_id": "1"}
    ]
    """

  Scenario: User's default language is English  (parents are visible)
    Given I am the user with group_id "51"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null, "language_id": null},
      {"id": "60", "team_only_contest": true, "parents": [{"title": "Chapter", "language_id": "1"}],
       "title": "Contest", "language_id": "1"},
      {"id": "70", "team_only_contest": false,
       "parents": [{"title": "Chapter", "language_id": "1"}, {"title": "Contest", "language_id": "1"}],
       "title": "Contest 2", "language_id": "1"}
    ]
    """

  Scenario: Empty result
    Given I am the user with group_id "61"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: User's default language is English  (parents are visible), limit=1
    Given I am the user with group_id "51"
    When I send a GET request to "/contests/administered?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null, "language_id": null}
    ]
    """

  Scenario: User's default language is English  (parents are visible), start from the second row, limit=1
    Given I am the user with group_id "51"
    When I send a GET request to "/contests/administered?from.title&from.id=50&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "60", "team_only_contest": true, "parents": [{"title": "Chapter", "language_id": "1"}],
       "title": "Contest", "language_id": "1"}
    ]
    """

  Scenario: User's default language is English  (parents are visible), inverse order
    Given I am the user with group_id "51"
    When I send a GET request to "/contests/administered?sort=-title,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "70", "team_only_contest": false,
       "parents": [{"title": "Chapter", "language_id": "1"}, {"title": "Contest", "language_id": "1"}],
       "title": "Contest 2", "language_id": "1"},
      {"id": "60", "team_only_contest": true, "parents": [{"title": "Chapter", "language_id": "1"}],
       "title": "Contest", "language_id": "1"},
      {"id": "50", "team_only_contest": false, "parents": [], "title": null, "language_id": null}
    ]
    """

  Scenario: Keeps parents with nil titles
    Given I am the user with group_id "71"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "80", "team_only_contest": false, "parents": [{"language_id": null, "title": null}],
       "title": null, "language_id": null},
      {"id": "90", "team_only_contest": false, "parents": [], "title": null, "language_id": null}
    ]
    """
