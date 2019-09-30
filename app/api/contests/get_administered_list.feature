Feature: Get the contests that the user has administration rights on (contestAdminList)
  Background:
    Given the database has the following table 'users':
      | id | login          | self_group_id | owned_group_id | default_language |
      | 1  | possesseur     | 21            | 22             | fr               |
      | 2  | owner          | 31            | 32             | en               |
      | 3  | administrateur | 41            | 42             | fr               |
      | 4  | admin          | 51            | 52             | en               |
      | 5  | guest          | 61            | 62             | en               |
      | 6  | panas          | 71            | 72             | uk               |
    And the database has the following table 'languages':
      | id | code |
      | 1  | en   |
      | 2  | fr   |
      | 3  | uk   |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
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
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_since | cached_grayed_access_since | cached_full_access_since | cached_solutions_access_since | creator_user_id |
      | 21       | 50      | null                        | null                       | null                     | 2018-05-29 06:38:38           | 4               |
      | 21       | 60      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 4               |
      | 21       | 70      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 4               |
      | 31       | 50      | null                        | null                       | null                     | 2018-05-29 06:38:38           | 4               |
      | 31       | 60      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 4               |
      | 31       | 70      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 4               |
      | 41       | 10      | 2018-05-29 06:38:38         | null                       | null                     | null                          | 4               |
      | 41       | 50      | null                        | null                       | null                     | 2018-05-29 06:38:38           | 4               |
      | 41       | 60      | null                        | 2018-05-29 06:38:38        | null                     | 2018-05-29 06:38:38           | 4               |
      | 41       | 70      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 4               |
      | 51       | 10      | null                        | 2018-05-29 06:38:38        | null                     | null                          | 4               |
      | 51       | 50      | null                        | null                       | null                     | 2018-05-29 06:38:38           | 4               |
      | 51       | 60      | null                        | null                       | 2018-05-29 06:38:38      | 2018-05-29 06:38:38           | 4               |
      | 51       | 70      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 4               |
      | 71       | 80      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 4               |
      | 71       | 90      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 4               |

  Scenario: User's default language is French (most parents are invisible)
    Given I am the user with id "1"
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
    Given I am the user with id "2"
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
    Given I am the user with id "3"
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
    Given I am the user with id "4"
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
    Given I am the user with id "5"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: User's default language is English  (parents are visible), limit=1
    Given I am the user with id "4"
    When I send a GET request to "/contests/administered?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null, "language_id": null}
    ]
    """

  Scenario: User's default language is English  (parents are visible), start from the second row, limit=1
    Given I am the user with id "4"
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
    Given I am the user with id "4"
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
    Given I am the user with id "6"
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
