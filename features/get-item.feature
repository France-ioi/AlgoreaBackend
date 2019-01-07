@wip
Feature: Get item for tree navigation

  Scenario: Get tree structure
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | iVersion |
      | 1  | jdoe   | 0        | 11          | 12           | 0        |
    And the database has the following table 'groups':
      | ID | sName      | sTextId | iGrade | sType     | iVersion |
      | 11 | jdoe       |         | -2     | UserAdmin | 0        |
      | 12 | jdoe-admin |         | -2     | UserAdmin | 0        |
      | 13 | Group B    |         | -2     | Class     | 0        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | iVersion |
      | 61 | 13            | 11           | 0        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 71 | 11              | 11           | 1       | 0        |
      | 72 | 12              | 12           | 1       | 0        |
      | 73 | 13              | 13           | 1       | 0        |
      | 74 | 13              | 11           | 0       | 0        |
    And the database has the following table 'items':
      | ID | bTeamsEditable | bNoScore | iVersion |
      | 23 | false          | false    | 0        |
      | 24 | false          | false    | 0        |
      | 25 | false          | false    | 0        |
      | 26 | false          | false    | 0        |
      | 41 | false          | false    | 0        |
      | 61 | false          | false    | 0        |
      | 62 | false          | false    | 0        |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
      | 43 | 13      | 23     | null            | true              | false                | false               | 0             | 0        |
    And the database has the following table 'items_items':
      | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
      | 54 | 23           | 24          | 1           | 0           | 0        |
      | 55 | 23           | 25          | 2           | 0           | 0        |
      | 56 | 23           | 26          | 3           | 0           | 0        |
      | 57 | 24           | 41          | 3           | 0           | 0        |
      | 58 | 26           | 61          | 3           | 0           | 0        |
      | 59 | 26           | 62          | 3           | 0           | 0        |
    And the database has the following table 'items_strings':
      | ID | idItem | idLanguage | sTitle    | iVersion |
      | 53 | 23     | 1          | Root      | 0        |
      | 54 | 24     | 1          | Chapter A | 0        |
      | 55 | 25     | 1          | Chapter B | 0        |
      | 56 | 26     | 2          | Chapter C | 0        |
      | 57 | 41     | 2          | Lesson 1  | 0        |
      | 58 | 61     | 2          | Lesson 2  | 0        |
      | 59 | 62     | 2          | Lesson 3  | 0        |
    And I am the user with ID "1"
    When I send a GET request to "/items/23"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
      "item_id":23,
      "title":"Root",
      "children":
      [{
      "item_id":24,"order":1,
      "title":"Chapter A",
      "children":[{"item_id":41,"order":3,"title":"Lesson 1"}]
      },{
      "item_id":25,
      "order":2,
      "title":"Chapter B"
      },{
      "item_id":26,
      "order":3,
      "title":"Chapter C",
      "children":[
      {"item_id":61,"order":3,"title":"Lesson 2"},
      {"item_id":62,"order":3,"title":"Lesson 3"}]
      }]
      }
      """
