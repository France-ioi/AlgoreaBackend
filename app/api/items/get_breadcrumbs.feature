Feature: Get item information for breadcrumb

Background:
  Given the database has the following table 'groups':
    | id | name    | text_id | grade | type     |
    | 11 | jdoe    |         | -2    | UserSelf |
    | 13 | Group B |         | -2    | Class    |
  And the database has the following table 'languages':
    | id | code |
    | 1  | en   |
    | 2  | fr   |
  And the database has the following table 'users':
    | login | group_id | default_language |
    | jdoe  | 11       | fr               |
  And the database has the following table 'items':
    | id | type     | default_language_id |
    | 21 | Root     | 1                   |
    | 22 | Category | 1                   |
    | 23 | Chapter  | 1                   |
    | 24 | Task     | 2                   |
  And the database has the following table 'items_strings':
    | id | item_id | language_id | title            |
    | 31 | 21      | 1           | Graph: Methods   |
    | 32 | 22      | 1           | DFS              |
    | 33 | 23      | 1           | Reduce Graph     |
    | 39 | 21      | 2           | Graphe: Methodes |
  And the database has the following table 'groups_groups':
    | id | parent_group_id | child_group_id |
    | 61 | 13              | 11             |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
    | 73 | 13                | 13             | 1       |
    | 74 | 13                | 11             | 0       |

Scenario: Full access on all breadcrumb
  Given the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 21      | content_with_descendants |
    | 13       | 22      | content_with_descendants |
    | 13       | 23      | content_with_descendants |
  And the database has the following table 'items_items':
    | id | parent_item_id | child_item_id | child_order | difficulty |
    | 51 | 21             | 22            | 1           | 0          |
    | 52 | 22             | 23            | 1           | 0          |
  And I am the user with id "11"
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

Scenario: 'Content' access on all breadcrumb
  Given the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated |
    | 13       | 21      | content            |
    | 13       | 22      | content            |
    | 13       | 23      | content            |
  And the database has the following table 'items_items':
    | id | parent_item_id | child_item_id | child_order | difficulty |
    | 51 | 21             | 22            | 1           | 0          |
    | 52 | 22             | 23            | 1           | 0          |
  And I am the user with id "11"
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

Scenario: Content access to all items except for last for which we have info access
  Given the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated |
    | 13       | 21      | content            |
    | 13       | 22      | content            |
    | 13       | 23      | info               |
  And the database has the following table 'items_items':
    | id | parent_item_id | child_item_id | child_order | difficulty |
    | 51 | 21             | 22            | 1           | 0          |
    | 52 | 22             | 23            | 1           | 0          |
  And I am the user with id "11"
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

