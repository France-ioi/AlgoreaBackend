Feature: Get item information for breadcrumb - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    | text_id | grade | type     |
      | 11 | jdoe    |         | -2    | UserSelf |
      | 13 | Group B |         | -2    | Class    |
    And the database has the following table 'users':
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'items':
      | id | teams_editable | no_score | type     |
      | 21 | false          | false    | Root     |
      | 22 | false          | false    | Category |
      | 23 | false          | false    | Chapter  |
      | 24 | false          | false    | Task     |
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

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one parent-child link missing), but user has full access to all
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 21      | content_with_descendants |
      | 13       | 22      | content_with_descendants |
      | 13       | 23      | content_with_descendants |
    And the database has the following table 'items_items':
      | id | parent_item_id | child_item_id | child_order | difficulty |
      | 52 | 22             | 23            | 1           | 0          |
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "The IDs chain is corrupt"

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one parent-child link missing at the end), but user has full access to all
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 21      | content_with_descendants |
      | 13       | 22      | content_with_descendants |
      | 13       | 23      | content_with_descendants |
      | 13       | 24      | content_with_descendants |
    And the database has the following table 'items_items':
      | id | parent_item_id | child_item_id | child_order | difficulty |
      | 52 | 21             | 22            | 1           | 0          |
      | 53 | 22             | 23            | 1           | 0          |
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/24/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "The IDs chain is corrupt"

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one item missing), and user has full access to all
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 21      | content_with_descendants |
      | 13       | 22      | content_with_descendants |
      | 13       | 24      | content_with_descendants |
    And the database has the following table 'items_items':
      | id | parent_item_id | child_item_id | child_order | difficulty |
      | 51 | 21             | 22            | 1           | 0          |
      | 52 | 22             | 23            | 1           | 0          |
      | 53 | 23             | 24            | 1           | 0          |
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/24/23/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item ids"

  Scenario: Should fail when the first item of hierarchy is not a root item, and user has full access to all
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 22      | content_with_descendants |
      | 13       | 23      | content_with_descendants |
    And the database has the following table 'items_items':
      | id | parent_item_id | child_item_id | child_order | difficulty |
      | 52 | 22             | 23            | 1           | 0          |
      | 53 | 23             | 24            | 1           | 0          |
    And I am the user with id "11"
    When I send a GET request to "/items/22/23/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "The IDs chain is corrupt"

  Scenario: Should fail when the user has 'info' access to middle element, 'content' access to the rest
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 13       | 21      | content            |
      | 13       | 22      | info               |
      | 13       | 23      | content            |
    And the database has the following table 'items_items':
      | id | parent_item_id | child_item_id | child_order | difficulty |
      | 52 | 22             | 23            | 1           | 0          |
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item ids"

  Scenario: Should fail when the user doesn't exist
    And I am the user with id "404"
    When I send a GET request to "/items/21/22/23/breadcrumbs"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Invalid ids
    And I am the user with id "11"
    When I send a GET request to "/items/11111111111111111111111111111/2222222222222222222222222222/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: '11111111111111111111111111111', param: 'ids')"

  Scenario: More than 10 ids
    And I am the user with id "11"
    When I send a GET request to "/items/1/2/3/4/5/6/7/8/9/10/11/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "No more than 10 ids expected"
