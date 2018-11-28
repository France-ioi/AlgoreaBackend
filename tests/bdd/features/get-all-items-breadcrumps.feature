@wip
Feature: Get item information for breadcrumb

Scenario: All items from a valid breadcrumb
Given the database has the following table 'users':
  | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | iVersion |
  |  1 |  jdoe  |        0 |          11 |           12 |        0 |
And the database has the following table 'groups':
  | ID |      sName | sTextId | iGrade |     sType | iVersion |
  | 11 |       jdoe |         |     -2 | UserAdmin |        0 |
  | 12 | jdoe-admin |         |     -2 | UserAdmin |        0 |
And the database has the following table 'items':
  | ID | bTeamsEditable | bNoScore | iVersion |
  | 21 |          false |    false |        0 |
  | 22 |          false |    false |        0 |
  | 23 |          false |    false |        0 |
And the database has the following table 'items_strings':
  | ID | idItem | idLanguage |           sTitle | iVersion |
  | 31 |     21 |          1 |   Graph: Methods |        0 |
  | 32 |     22 |          1 |              DFS |        0 |
  | 33 |     23 |          1 |     Reduce Graph |        0 |
  | 39 |     21 |          2 | Graphe: Methodes |        0 |
And the database has the following table 'groups_items':
  | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedGrayedAccess | idUserCreated | iVersion |
  | 41 |      13 |     21 |      2010-01-01 |              true |               false |            0  |        0 |
  | 42 |      13 |     22 |            null |              true |               false |            0  |        0 |
  | 43 |      13 |     23 |            null |              true |               false |            0  |        0 |
  | 44 |      11 |     21 |            null |              true |               false |            0  |        0 |
  | 45 |      11 |     22 |            null |              true |               false |            0  |        0 |
  | 46 |      11 |     23 |            null |              true |               false |            0  |        0 |
And the database has the following table 'items_items':
  | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
  | 51 |           21 |          22 |           1 |           0 |        0 |
  | 52 |           22 |          23 |           1 |           0 |        0 |
And the database has the following table 'groups_groups':
  | ID | idGroupParent | idGroupChild | iVersion |
  | 61 |            13 |           11 |        0 |
And the database has the following table 'groups_ancestors':
  | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
  | 71 |              11 |           11 |       1 |        0 |
  | 72 |              12 |           12 |       1 |        0 |
  | 73 |              13 |           13 |       1 |        0 |
  | 74 |              13 |           11 |       0 |        0 |
And I am the user with ID "1" and my preferred language is "1"
When I send a GET request to "/items/?ids=21,22,23"
Then the response code should be 200
And the response body should be, in JSON:
  """
  [
    { "id": 31, "item_id": 21, "language_id": 1, "title": "Graph: Methods" }
    { "id": 32, "item_id": 22, "language_id": 1, "title": "DFS"}
    { "id": 33, "item_id": 23, "language_id": 1, "title": "Reduce Graph" }
  ]
  """
