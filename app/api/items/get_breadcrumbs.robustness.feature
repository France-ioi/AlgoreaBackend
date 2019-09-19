Feature: Get item information for breadcrumb - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | temp_user | group_self_id | group_owned_id | version |
      | 1  | jdoe  | 0         | 11            | 12             | 0       |
    And the database has the following table 'groups':
      | id | name       | text_id | grade | type      | version |
      | 11 | jdoe       |         | -2    | UserAdmin | 0       |
      | 12 | jdoe-admin |         | -2    | UserAdmin | 0       |
      | 13 | Group B    |         | -2    | Class     | 0       |
    And the database has the following table 'items':
      | id | teams_editable | no_score | version | type     |
      | 21 | false          | false    | 0       | Root     |
      | 22 | false          | false    | 0       | Category |
      | 23 | false          | false    | 0       | Chapter  |
      | 24 | false          | false    | 0       | Task     |
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

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one parent-child link missing), but user has full access to all
    Given the database has the following table 'groups_items':
      | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
      | 41 | 13       | 21      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 42 | 13       | 22      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 43 | 13       | 23      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    And the database has the following table 'items_items':
      | id | item_parent_id | item_child_id | child_order | difficulty | version |
      | 52 | 22             | 23            | 1           | 0          | 0       |
    And I am the user with id "1"
    When I send a GET request to "/items/21/22/23/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "The IDs chain is corrupt"

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one parent-child link missing at the end), but user has full access to all
    Given the database has the following table 'groups_items':
      | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
      | 41 | 13       | 21      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 42 | 13       | 22      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 43 | 13       | 23      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 44 | 13       | 24      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    And the database has the following table 'items_items':
      | id | item_parent_id | item_child_id | child_order | difficulty | version |
      | 52 | 21             | 22            | 1           | 0          | 0       |
      | 53 | 22             | 23            | 1           | 0          | 0       |
    And I am the user with id "1"
    When I send a GET request to "/items/21/22/23/24/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "The IDs chain is corrupt"

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one item missing), and user has full access to all
    Given the database has the following table 'groups_items':
      | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
      | 41 | 13       | 21      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 42 | 13       | 22      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 44 | 13       | 24      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    And the database has the following table 'items_items':
      | id | item_parent_id | item_child_id | child_order | difficulty | version |
      | 51 | 21             | 22            | 1           | 0          | 0       |
      | 52 | 22             | 23            | 1           | 0          | 0       |
      | 53 | 23             | 24            | 1           | 0          | 0       |
    And I am the user with id "1"
    When I send a GET request to "/items/21/22/24/23/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item ids"

  Scenario: Should fail when the first item of hierarchy is not a root item, and user has full access to all
    Given the database has the following table 'groups_items':
      | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
      | 42 | 13       | 22      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 44 | 13       | 23      | 2017-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    And the database has the following table 'items_items':
      | id | item_parent_id | item_child_id | child_order | difficulty | version |
      | 52 | 22             | 23            | 1           | 0          | 0       |
      | 53 | 23             | 24            | 1           | 0          | 0       |
    And I am the user with id "1"
    When I send a GET request to "/items/22/23/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "The IDs chain is corrupt"

  Scenario: Should fail when the user has greyed access to middle element, partial access to the rest
    Given the database has the following table 'groups_items':
      | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
      | 41 | 13       | 21      | 2037-05-29 06:38:38     | 2017-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
      | 42 | 13       | 22      | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2017-05-29 06:38:38       | 0               | 0       |
      | 43 | 13       | 23      | 2037-05-29 06:38:38     | 2017-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    And the database has the following table 'items_items':
      | id | item_parent_id | item_child_id | child_order | difficulty | version |
      | 52 | 22             | 23            | 1           | 0          | 0       |
    And I am the user with id "1"
    When I send a GET request to "/items/21/22/23/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item ids"

  Scenario: Should fail when the user doesn't exist
    And I am the user with id "10"
    When I send a GET request to "/items/21/22/23/breadcrumbs"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Invalid ids
    And I am the user with id "1"
    When I send a GET request to "/items/11111111111111111111111111111/2222222222222222222222222222/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: '11111111111111111111111111111', param: 'ids')"

  Scenario: More than 10 ids
    And I am the user with id "1"
    When I send a GET request to "/items/1/2/3/4/5/6/7/8/9/10/11/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "No more than 10 ids expected"
