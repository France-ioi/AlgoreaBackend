@wip
Feature: Get item information for breadcrumb - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | iVersion |
      | 1  | jdoe   | 0        | 11          | 12           | 0        |
    And the database has the following table 'groups':
      | ID | sName      | sTextId | iGrade | sType     | iVersion |
      | 11 | jdoe       |         | -2     | UserAdmin | 0        |
      | 12 | jdoe-admin |         | -2     | UserAdmin | 0        |
      | 13 | Group B    |         | -2     | Class     | 0        |
    And the database has the following table 'items':
      | ID | bTeamsEditable | bNoScore | iVersion | sType    |
      | 21 | false          | false    | 0        | Root     |
      | 22 | false          | false    | 0        | Category |
      | 23 | false          | false    | 0        | Chapter  |
      | 24 | false          | false    | 0        | Task     |
    And the database has the following table 'items_strings':
      | ID | idItem | idLanguage | sTitle           | iVersion |
      | 31 | 21     | 1          | Graph: Methods   | 0        |
      | 32 | 22     | 1          | DFS              | 0        |
      | 33 | 23     | 1          | Reduce Graph     | 0        |
      | 39 | 21     | 2          | Graphe: Methodes | 0        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | iVersion |
      | 61 | 13            | 11           | 0        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 71 | 11              | 11           | 1       | 0        |
      | 72 | 12              | 12           | 1       | 0        |
      | 73 | 13              | 13           | 1       | 0        |
      | 74 | 13              | 11           | 0       | 0        |

  Scenario: Full access on all breadcrumb
  Given the database has the following table 'groups_items':
    | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
    | 41 | 13      | 21     | 2010-01-01      | true              | false                | false               | 0             | 0        |
    | 42 | 13      | 22     | null            | true              | false                | false               | 0             | 0        |
    | 43 | 13      | 23     | null            | true              | false                | false               | 0             | 0        |
  And the database has the following table 'items_items':
    | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
    | 51 | 21           | 22          | 1           | 0           | 0        |
    | 52 | 22           | 23          | 1           | 0           | 0        |
  And I am the user with ID "1"
  When I send a GET request to "/items/?ids=21,22,23"
  Then the response code should be 200
  And the response body should be, in JSON:
  """
  [
  { "item_id": 21, "language_id": 1, "title": "Graph: Methods" },
  { "item_id": 22, "language_id": 1, "title": "DFS" },
  { "item_id": 23, "language_id": 1, "title": "Reduce Graph" },
  { "item_id": 21, "language_id": 2, "title": "Graphe: Methodes" }
  ]
  """

Scenario: Should fail when breadcrumb hierarchy is corrupt (one parent-child link missing), but user has full access to all
  Given the database has the following table 'groups_items':
    | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
    | 41 | 13      | 21     | 2010-01-01      | true              | false                | false               | 0             | 0        |
    | 42 | 13      | 22     | null            | true              | false                | false               | 0             | 0        |
    | 43 | 13      | 23     | null            | true              | false                | false               | 0             | 0        |
  And the database has the following table 'items_items':
    | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
    | 52 | 22           | 23          | 1           | 0           | 0        |
  And I am the user with ID "1"
  When I send a GET request to "/items/?ids=21,22,23"
  Then the response code should be 400
  And the response error message should contain "The IDs chain is corrupt"

Scenario: Should fail when breadcrumb hierarchy is corrupt (one parent-child link missing at the end), but user has full access to all
  Given the database has the following table 'groups_items':
    | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
    | 41 | 13      | 21     | 2010-01-01      | true              | false                | false               | 0             | 0        |
    | 42 | 13      | 22     | null            | true              | false                | false               | 0             | 0        |
    | 43 | 13      | 23     | null            | true              | false                | false               | 0             | 0        |
    | 44 | 13      | 24     | null            | true              | false                | false               | 0             | 0        |
  And the database has the following table 'items_items':
    | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
    | 52 | 21           | 22          | 1           | 0           | 0        |
    | 53 | 22           | 23          | 1           | 0           | 0        |
  And I am the user with ID "1"
  When I send a GET request to "/items/?ids=21,22,23,24"
  Then the response code should be 400
  And the response error message should contain "The IDs chain is corrupt"

Scenario: Should fail when breadcrumb hierarchy is corrupt (one item missing), and user has full access to all
  Given the database has the following table 'groups_items':
    | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
    | 41 | 13      | 21     | 2010-01-01      | true              | false                | false               | 0             | 0        |
    | 42 | 13      | 22     | null            | true              | false                | false               | 0             | 0        |
    | 44 | 13      | 24     | null            | true              | false                | false               | 0             | 0        |
  And the database has the following table 'items_items':
    | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
    | 51 | 21           | 22          | 1           | 0           | 0        |
    | 52 | 22           | 23          | 1           | 0           | 0        |
    | 53 | 23           | 24          | 1           | 0           | 0        |
  And I am the user with ID "1"
  When I send a GET request to "/items/?ids=21,22,24,23"
  Then the response code should be 403
  And the response error message should contain "Insufficient access on given item ids"

Scenario: Should fail when the first item of hierarchy is not a root item, and user has full access to all
  Given the database has the following table 'groups_items':
    | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
    | 42 | 13      | 22     | null            | true              | false                | false               | 0             | 0        |
    | 44 | 13      | 23     | null            | true              | false                | false               | 0             | 0        |
  And the database has the following table 'items_items':
    | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
    | 52 | 22           | 23          | 1           | 0           | 0        |
    | 53 | 23           | 24          | 1           | 0           | 0        |
  And I am the user with ID "1"
  When I send a GET request to "/items/?ids=22,23"
  Then the response code should be 400
  And the response error message should contain "The IDs chain is corrupt"

Scenario: Should fail when the user has greyed access to middle element, partial access to the rest
  Given the database has the following table 'groups_items':
    | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
    | 41 | 13      | 21     | 2010-01-01      | false             | true                 | false               | 0             | 0        |
    | 42 | 13      | 22     | null            | false             | false                | true                | 0             | 0        |
    | 43 | 13      | 23     | null            | false             | true                 | false               | 0             | 0        |
  And the database has the following table 'items_items':
    | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
    | 52 | 22           | 23          | 1           | 0           | 0        |
  And I am the user with ID "1"
  When I send a GET request to "/items/?ids=21,22,23"
  Then the response code should be 403
  And the response error message should contain "Insufficient access on given item ids"
