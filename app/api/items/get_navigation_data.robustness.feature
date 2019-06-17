Feature: Get item for tree navigation - robustness
Background:
  Given the database has the following table 'users':
    | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | iVersion |
    | 1  | jdoe   | 0        | 11          | 12           | 0        |
    | 2  | guest  | 0        | 404         | 404          | 0        |
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
    | ID  | sType    | bTeamsEditable | bNoScore | idItemUnlocked | bTransparentFolder | iVersion |
    | 190 | Category | false          | false    | 1234,2345      | true               | 0        |
    | 200 | Category | false          | false    | 1234,2345      | true               | 0        |
  And the database has the following table 'groups_items':
    | ID | idGroup | idItem | sCachedFullAccessDate | sCachedPartialAccessDate | sCachedGrayedAccessDate | idUserCreated | iVersion |
    | 42 | 13      | 190    | 2037-05-29T06:38:38Z  | 2037-05-29T06:38:38Z     | 2037-05-29T06:38:38Z    | 0             | 0        |
    | 43 | 13      | 200    | 2017-05-29T06:38:38Z  | 2017-05-29T06:38:38Z     | 2017-05-29T06:38:38Z    | 0             | 0        |
  And the database has the following table 'items_strings':
    | ID | idItem | idLanguage | sTitle     | iVersion |
    | 53 | 200    | 1          | Category 1 | 0        |
  And the database has the following table 'users_items':
    | ID | idUser | idItem | iScore | nbSubmissionsAttempts | bValidated  | bFinished | bKeyObtained | sStartDate           | sFinishDate          | sValidationDate      | iVersion |
    | 1  | 1      | 200    | 12345  | 10                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |

  Scenario: Should fail when the user doesn't have access to the root item
    Given I am the user with ID "1"
    When I send a GET request to "/items/190/as-nav-tree"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with ID "10"
    When I send a GET request to "/items/190/as-nav-tree"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't have access to the root item (for a user with a non-existent group)
    Given I am the user with ID "2"
    When I send a GET request to "/items/200/as-nav-tree"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Should fail when the root item doesn't exist
    Given I am the user with ID "1"
    When I send a GET request to "/items/404/as-nav-tree"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Invalid item_id
    Given I am the user with ID "1"
    When I send a GET request to "/items/abc/as-nav-tree"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
