Feature: Get item breadcrumbs

Background:
  Given the database has the following table "groups":
    | id | name    | type  | root_activity_id | root_skill_id |
    | 11 | jdoe    | User  | 22               | null          |
    | 13 | Group B | Class | 21               | null          |
    | 14 | Team B  | Team  | 23               | 25            |
    | 15 | Group C | Class | 21               | null          |
  And the database has the following table "languages":
    | tag |
    | en  |
    | fr  |
  And the database has the following user:
    | group_id | login | default_language |
    | 11       | jdoe  | fr               |
  And the database has the following table "items":
    | id | type    | default_language_tag | allows_multiple_attempts |
    | 21 | Task    | en                   | 0                        |
    | 22 | Chapter | en                   | 1                        |
    | 23 | Chapter | en                   | 1                        |
    | 24 | Task    | fr                   | 0                        |
    | 25 | Chapter | en                   | 1                        |
    | 26 | Chapter | en                   | 1                        |
  And the database has the following table "items_strings":
    | item_id | language_tag | title            |
    | 21      | en           | Graph: Methods   |
    | 21      | fr           | Graphe: Methodes |
    | 22      | en           | DFS              |
    | 23      | en           | Reduce Graph     |
    | 25      | en           | BFS              |
    | 26      | en           | Trees            |
  And the database has the following table "groups_groups":
    | parent_group_id | child_group_id |
    | 13              | 11             |
    | 14              | 11             |
    | 15              | 14             |
  And the groups ancestors are computed

Scenario: Full access on all the breadcrumbs (as a user)
  Given the database has the following table "permissions_generated":
    | group_id | item_id | can_view_generated       |
    | 13       | 21      | content_with_descendants |
    | 13       | 22      | content_with_descendants |
    | 13       | 23      | content_with_descendants |
  And the database has the following table "items_items":
    | parent_item_id | child_item_id | child_order |
    | 21             | 22            | 1           |
    | 22             | 23            | 1           |
  And the database has the following table "attempts":
    | participant_id | id | parent_attempt_id | root_item_id |
    | 11             | 0  | null              | null         |
    | 11             | 1  | 0                 | 22           |
    | 11             | 2  | 0                 | 22           |
  And the database has the following table "results":
    | participant_id | attempt_id | item_id | started_at          |
    | 11             | 0          | 21      | 2019-05-30 11:00:00 |
    | 11             | 1          | 22      | 2019-05-30 11:00:00 |
    | 11             | 2          | 22      | 2019-05-29 11:00:00 |
    | 11             | 1          | 23      | 2019-05-29 11:00:00 |
    | 11             | 2          | 23      | 2019-05-30 11:00:00 |
  And I am the user with id "11"
  When I send a GET request to "/items/21/22/23/breadcrumbs?attempt_id=1"
  Then the response code should be 200
  And the response body should be, in JSON:
  """
  [
    { "item_id": "21", "type": "Task", "language_tag": "fr", "title": "Graphe: Methodes", "attempt_id": "0" },
    { "item_id": "22", "type": "Chapter", "language_tag": "en", "title": "DFS", "attempt_id": "1", "attempt_order": 2 },
    { "item_id": "23", "type": "Chapter", "language_tag": "en", "title": "Reduce Graph", "attempt_id": "1", "attempt_order": 1 }
  ]
  """

Scenario: 'Content' access on all the breadcrumbs (as a team)
  Given the database has the following table "permissions_generated":
    | group_id | item_id | can_view_generated |
    | 14       | 21      | content            |
    | 14       | 22      | content            |
    | 14       | 23      | content            |
  And the database has the following table "items_items":
    | parent_item_id | child_item_id | child_order |
    | 21             | 22            | 1           |
    | 22             | 23            | 1           |
  And the database has the following table "attempts":
    | participant_id | id | parent_attempt_id | root_item_id |
    | 14             | 0  | null              | null         |
    | 14             | 1  | 0                 | 22           |
    | 14             | 2  | 0                 | 22           |
  And the database has the following table "results":
    | participant_id | attempt_id | item_id | started_at          |
    | 14             | 0          | 21      | 2019-05-30 11:00:00 |
    | 14             | 1          | 22      | 2019-05-29 11:00:00 |
    | 14             | 1          | 23      | 2019-05-30 11:00:00 |
    | 14             | 2          | 22      | 2019-05-30 11:00:00 |
    | 14             | 2          | 23      | 2019-05-29 11:00:00 |
  And I am the user with id "11"
  When I send a GET request to "/items/21/22/23/breadcrumbs?as_team_id=14&attempt_id=1"
  Then the response code should be 200
  And the response body should be, in JSON:
    """
    [
      { "item_id": "21", "type": "Task", "language_tag": "fr", "title": "Graphe: Methodes", "attempt_id": "0" },
      { "item_id": "22", "type": "Chapter", "language_tag": "en", "title": "DFS", "attempt_id": "1", "attempt_order": 1 },
      { "item_id": "23", "type": "Chapter", "language_tag": "en", "title": "Reduce Graph", "attempt_id": "1", "attempt_order": 2 }
    ]
    """

Scenario: Content access to all items except for last for which we have info access
  Given the database has the following table "permissions_generated":
    | group_id | item_id | can_view_generated |
    | 13       | 21      | content            |
    | 13       | 22      | content            |
    | 13       | 23      | info               |
  And the database has the following table "items_items":
    | parent_item_id | child_item_id | child_order |
    | 21             | 22            | 1           |
    | 22             | 23            | 1           |
  And the database has the following table "attempts":
    | participant_id | id | parent_attempt_id | root_item_id |
    | 11             | 0  | null              | null         |
    | 11             | 1  | 0                 | 22           |
  And the database has the following table "results":
    | participant_id | attempt_id | item_id | started_at          |
    | 11             | 0          | 21      | 2019-05-30 11:00:00 |
    | 11             | 1          | 22      | 2019-05-30 11:00:00 |
  And I am the user with id "11"
  When I send a GET request to "/items/21/22/23/breadcrumbs?parent_attempt_id=1"
  Then the response code should be 200
  And the response body should be, in JSON:
    """
    [
      { "item_id": "21", "type": "Task", "language_tag": "fr", "title": "Graphe: Methodes", "attempt_id": "0" },
      { "item_id": "22", "type": "Chapter", "language_tag": "en", "title": "DFS", "attempt_id": "1", "attempt_order": 1 },
      { "item_id": "23", "type": "Chapter", "language_tag": "en", "title": "Reduce Graph" }
    ]
    """

  Scenario: Allows the first item to be root_activity_id of some participant's ancestor
    Given the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 14       | 23      | info               |
    And I am the user with id "11"
    When I send a GET request to "/items/23/breadcrumbs?parent_attempt_id=0&as_team_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      { "item_id": "23", "type": "Chapter", "language_tag": "en", "title": "Reduce Graph" }
    ]
    """

  Scenario: Allows the first item to be root_skill_id of some participant's ancestor
    Given the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 14       | 25      | info               |
    And I am the user with id "11"
    When I send a GET request to "/items/25/breadcrumbs?parent_attempt_id=0&as_team_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      { "item_id": "25", "type": "Chapter", "language_tag": "en", "title": "BFS" }
    ]
    """

  Scenario: Allows the first item to be root_activity_id of a managed group
    Given the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 14       | 26      | info               |
    And the database has the following table "group_managers":
      | manager_id | group_id |
      | 14         | 16       |
    And the database table "groups" also has the following row:
      | id | name    | root_activity_id |
      | 16 | Managed | 26               |
    And the groups ancestors are computed
    And I am the user with id "11"
    When I send a GET request to "/items/26/breadcrumbs?parent_attempt_id=0&as_team_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      { "item_id": "26", "type": "Chapter", "language_tag": "en", "title": "Trees" }
    ]
    """

  Scenario: Allows the first item to be root_skill_id of a managed group
    Given the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 14       | 26      | info               |
    And the database has the following table "group_managers":
      | manager_id | group_id |
      | 14         | 16       |
    And the database table "groups" also has the following row:
      | id | name    | root_skill_id |
      | 16 | Managed | 26            |
    And the groups ancestors are computed
    And I am the user with id "11"
    When I send a GET request to "/items/26/breadcrumbs?parent_attempt_id=0&as_team_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      { "item_id": "26", "type": "Chapter", "language_tag": "en", "title": "Trees" }
    ]
    """

  Scenario: Allows the first item to be root_activity_id of a group managed by the participant's ancestor
    Given the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 14       | 26      | info               |
    And the database has the following table "group_managers":
      | manager_id | group_id |
      | 15         | 16       |
    And the database table "groups" also has the following rows:
      | id | name     | root_activity_id |
      | 16 | Ancestor | null             |
      | 17 | Managed  | 26               |
    And the database table "groups_groups" also has the following row:
      | parent_group_id | child_group_id |
      | 16              | 17             |
    And the groups ancestors are computed
    And I am the user with id "11"
    When I send a GET request to "/items/26/breadcrumbs?parent_attempt_id=0&as_team_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      { "item_id": "26", "type": "Chapter", "language_tag": "en", "title": "Trees" }
    ]
    """

  Scenario: Allows the first item to be root_skill_id of a group managed by the participant's ancestor
    Given the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 14       | 26      | info               |
    And the database has the following table "group_managers":
      | manager_id | group_id |
      | 15         | 16       |
    And the database table "groups" also has the following rows:
      | id | name     | root_skill_id |
      | 16 | Ancestor | null          |
      | 17 | Managed  | 26            |
    And the database table "groups_groups" also has the following row:
      | parent_group_id | child_group_id |
      | 16              | 17             |
    And the groups ancestors are computed
    And I am the user with id "11"
    When I send a GET request to "/items/26/breadcrumbs?parent_attempt_id=0&as_team_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      { "item_id": "26", "type": "Chapter", "language_tag": "en", "title": "Trees" }
    ]
    """
