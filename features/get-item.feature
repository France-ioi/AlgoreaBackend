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
      | ID  | bTeamsEditable | bNoScore | iVersion |
      | 200 | false          | false    | 0        |
      | 210 | false          | false    | 0        |
      | 220 | false          | false    | 0        |
      | 230 | false          | false    | 0        |
      | 211 | false          | false    | 0        |
      | 231 | false          | false    | 0        |
      | 232 | false          | false    | 0        |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
      | 43 | 13      | 200    | null            | true              | false                | false               | 0             | 0        |
    And the database has the following table 'items_items':
      | ID | idItemParent | idItemChild | iChildOrder | iDifficulty | iVersion |
      | 54 | 200          | 210         | 1           | 0           | 0        |
      | 55 | 200          | 220         | 2           | 0           | 0        |
      | 56 | 200          | 230         | 3           | 0           | 0        |
      | 57 | 210          | 211         | 1           | 0           | 0        |
      | 58 | 230          | 231         | 1           | 0           | 0        |
      | 59 | 230          | 232         | 2           | 0           | 0        |
    And the database has the following table 'items_ancestors':
      | ID | idItemAncestor | idItemChild |
      | 51 | 200            | 210         |
      | 52 | 200            | 220         |
      | 53 | 200            | 230         |
      | 54 | 200            | 211         |
      | 55 | 200            | 231         |
      | 56 | 200            | 232         |
      | 57 | 210            | 211         |
      | 58 | 230            | 231         |
      | 59 | 230            | 232         |
    And the database has the following table 'items_strings':
      | ID | idItem | idLanguage | sTitle    | iVersion |
      | 53 | 200    | 1          | Root      | 0        |
      | 54 | 210    | 1          | Chapter A | 0        |
      | 55 | 220    | 1          | Chapter B | 0        |
      | 56 | 230    | 1          | Chapter C | 0        |
      | 57 | 211    | 1          | Lesson 1  | 0        |
      | 58 | 231    | 1          | Lesson 2  | 0        |
      | 59 | 232    | 1          | Lesson 3  | 0        |
    And I am the user with ID "1"
    When I send a GET request to "/items/200"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "item_id": 200,
        "title": "Root",
        "children": [
          {
            "item_id": 210,
            "order": 1,
            "title": "Chapter A",
            "children": [{ "item_id": 211, "order": 1, "title": "Lesson 1"}]
          },
          { "item_id": 220, "order": 2, "title": "Chapter B"},
          {
            "item_id": 230,
            "order": 3,
            "title": "Chapter C",
            "children": [
              {"item_id": 231, "order": 1, "title": "Lesson 2"},
              {"item_id": 232, "order": 2, "title": "Lesson 3"}
            ]
          }]
        }
      """
