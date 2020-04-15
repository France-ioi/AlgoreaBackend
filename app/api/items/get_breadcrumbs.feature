Feature: Get item breadcrumbs

Background:
  Given the database has the following table 'groups':
    | id | name    | text_id | grade | type  |
    | 11 | jdoe    |         | -2    | User  |
    | 13 | Group B |         | -2    | Class |
  And the database has the following table 'languages':
    | tag |
    | en  |
    | fr  |
  And the database has the following table 'users':
    | login | group_id | default_language |
    | jdoe  | 11       | fr               |
  And the database has the following table 'items':
    | id | type    | default_language_tag | is_root |
    | 21 | Course  | en                   | 1       |
    | 22 | Chapter | en                   | 1       |
    | 23 | Chapter | en                   | 0       |
    | 24 | Task    | fr                   | 0       |
  And the database has the following table 'items_strings':
    | item_id | language_tag | title            |
    | 21      | en           | Graph: Methods   |
    | 22      | en           | DFS              |
    | 23      | en           | Reduce Graph     |
    | 21      | fr           | Graphe: Methodes |
  And the database has the following table 'groups_groups':
    | parent_group_id | child_group_id |
    | 13              | 11             |
  And the groups ancestors are computed

Scenario: Full access on all breadcrumb
  Given the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 21      | content_with_descendants |
    | 13       | 22      | content_with_descendants |
    | 13       | 23      | content_with_descendants |
  And the database has the following table 'items_items':
    | parent_item_id | child_item_id | child_order |
    | 21             | 22            | 1           |
    | 22             | 23            | 1           |
  And I am the user with id "11"
  When I send a GET request to "/items/21/22/23/breadcrumbs"
  Then the response code should be 200
  And the response body should be, in JSON:
  """
  [
    { "item_id": "21", "language_tag": "fr", "title": "Graphe: Methodes" },
    { "item_id": "22", "language_tag": "en", "title": "DFS" },
    { "item_id": "23", "language_tag": "en", "title": "Reduce Graph" }
  ]
  """

Scenario: 'Content' access on all breadcrumb
  Given the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated |
    | 13       | 21      | content            |
    | 13       | 22      | content            |
    | 13       | 23      | content            |
  And the database has the following table 'items_items':
    | parent_item_id | child_item_id | child_order |
    | 21             | 22            | 1           |
    | 22             | 23            | 1           |
  And I am the user with id "11"
  When I send a GET request to "/items/21/22/23/breadcrumbs"
  Then the response code should be 200
  And the response body should be, in JSON:
    """
    [
      { "item_id": "21", "language_tag": "fr", "title": "Graphe: Methodes" },
      { "item_id": "22", "language_tag": "en", "title": "DFS" },
      { "item_id": "23", "language_tag": "en", "title": "Reduce Graph" }
    ]
    """

Scenario: Content access to all items except for last for which we have info access
  Given the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated |
    | 13       | 21      | content            |
    | 13       | 22      | content            |
    | 13       | 23      | info               |
  And the database has the following table 'items_items':
    | parent_item_id | child_item_id | child_order |
    | 21             | 22            | 1           |
    | 22             | 23            | 1           |
  And I am the user with id "11"
  When I send a GET request to "/items/21/22/23/breadcrumbs"
  Then the response code should be 200
  And the response body should be, in JSON:
    """
    [
      { "item_id": "21", "language_tag": "fr", "title": "Graphe: Methodes" },
      { "item_id": "22", "language_tag": "en", "title": "DFS" },
      { "item_id": "23", "language_tag": "en", "title": "Reduce Graph" }
    ]
    """

