Feature: Get item information for breadcrumb

Background:
  Given the database has the following table 'users':
    | id | login | group_self_id | group_owned_id | default_language |
    | 1  | jdoe  | 11            | 12             | fr               |
  Given the database has the following table 'languages':
    | id | code |
    | 1  | en   |
    | 2  | fr   |
  And the database has the following table 'groups':
    | id | name       | text_id | grade | type      | version |
    | 11 | jdoe       |         | -2    | UserAdmin | 0       |
    | 12 | jdoe-admin |         | -2    | UserAdmin | 0       |
    | 13 | Group B    |         | -2    | Class     | 0       |
  And the database has the following table 'items':
    | id | type     | default_language_id |
    | 21 | Root     | 1                   |
    | 22 | Category | 1                   |
    | 23 | Chapter  | 1                   |
    | 24 | Task     | 2                   |
  And the database has the following table 'items_strings':
    | id | item_id | language_id | title            | version |
    | 31 | 21      | 1           | Graph: Methods   | 0       |
    | 32 | 22      | 1           | DFS              | 0       |
    | 33 | 23      | 1           | Reduce Graph     | 0       |
    | 39 | 21      | 2           | Graphe: Methodes | 0       |
  And the database has the following table 'groups_groups':
    | id | group_parent_id | group_child_id | version |
    | 61 | 13              | 11             | 0       |
  And the database has the following table 'groups_ancestors':
    | id | group_ancestor_id | group_child_id | is_self | version |
    | 71 | 11                | 11             | 1       | 0       |
    | 72 | 12                | 12             | 1       | 0       |
    | 73 | 13                | 13             | 1       | 0       |
    | 74 | 13                | 11             | 0       | 0       |

Scenario: Full access on all breadcrumb
  Given the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
    | 41 | 13       | 21      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    | 42 | 13       | 22      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    | 43 | 13       | 23      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
  And the database has the following table 'items_items':
    | id | item_parent_id | item_child_id | child_order | difficulty | version |
    | 51 | 21             | 22            | 1           | 0          | 0       |
    | 52 | 22             | 23            | 1           | 0          | 0       |
  And I am the user with id "1"
  When I send a GET request to "/items/21/22/23/breadcrumbs"
  Then the response code should be 200
  And the response body should be, in JSON:
  """
  [
    { "item_id": "21", "language_id": "2", "title": "Graphe: Methodes" },
    { "item_id": "22", "language_id": "1", "title": "DFS" },
    { "item_id": "23", "language_id": "1", "title": "Reduce Graph" }
  ]
  """

Scenario: Partial access on all breadcrumb
  Given the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
    | 41 | 13       | 21      | 2037-05-29 06:38:38     | 2017-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    | 42 | 13       | 22      | 2037-05-29 06:38:38     | 2017-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    | 43 | 13       | 23      | 2037-05-29 06:38:38     | 2017-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
  And the database has the following table 'items_items':
    | id | item_parent_id | item_child_id | child_order | difficulty | version |
    | 51 | 21             | 22            | 1           | 0          | 0       |
    | 52 | 22             | 23            | 1           | 0          | 0       |
  And I am the user with id "1"
  When I send a GET request to "/items/21/22/23/breadcrumbs"
  Then the response code should be 200
  And the response body should be, in JSON:
    """
    [
      { "item_id": "21", "language_id": "2", "title": "Graphe: Methodes" },
      { "item_id": "22", "language_id": "1", "title": "DFS" },
      { "item_id": "23", "language_id": "1", "title": "Reduce Graph" }
    ]
    """

Scenario: Partial access to all items except for last which is greyed
  Given the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
    | 41 | 13       | 21      | 2037-05-29 06:38:38     | 2017-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    | 42 | 13       | 22      | 2037-05-29 06:38:38     | 2017-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    | 43 | 13       | 23      | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2017-05-29 06:38:38       | 0               | 0       |
  And the database has the following table 'items_items':
    | id | item_parent_id | item_child_id | child_order | difficulty | version |
    | 51 | 21             | 22            | 1           | 0          | 0       |
    | 52 | 22             | 23            | 1           | 0          | 0       |
  And I am the user with id "1"
  When I send a GET request to "/items/21/22/23/breadcrumbs"
  Then the response code should be 200
  And the response body should be, in JSON:
    """
    [
      { "item_id": "21", "language_id": "2", "title": "Graphe: Methodes" },
      { "item_id": "22", "language_id": "1", "title": "DFS" },
      { "item_id": "23", "language_id": "1", "title": "Reduce Graph" }
    ]
    """

