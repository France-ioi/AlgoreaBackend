Feature: Change item access rights for a group
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  |
      | 2  | user   | 23          | 24           | John        | Doe       |
      | 3  | jane   | 31          | 32           | Jane        | Doe       |
    And the database has the following table 'groups':
      | ID | sName       | sType     |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
      | 23 | user        | UserSelf  |
      | 24 | user-admin  | UserAdmin |
      | 25 | some class  | Class     |
      | 31 | jane        | UserSelf  |
      | 32 | jane-admin  | UserAdmin |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
      | 22              | 23           | 0       |
      | 22              | 31           | 0       |
      | 23              | 23           | 1       |
      | 24              | 24           | 1       |
      | 25              | 23           | 0       |
      | 25              | 25           | 1       |
      | 31              | 31           | 1       |
      | 32              | 32           | 1       |
    And the database has the following table 'items':
      | ID  |
      | 100 |
      | 101 |
      | 102 |
      | 103 |
    And the database has the following table 'items_items':
      | idItemParent | idItemChild | bAlwaysVisible | bAccessRestricted |
      | 100          | 101         | true           | true              |
      | 101          | 102         | false          | false             |
      | 102          | 103         | false          | false             |
    And the database has the following table 'items_ancestors':
      | idItemAncestor | idItemChild |
      | 100            | 101         |
      | 100            | 102         |
      | 100            | 103         |
      | 101            | 102         |
      | 101            | 103         |
      | 102            | 103         |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sFullAccessDate | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bManagerAccess | bCachedManagerAccess | sAccessReason                                  |
      | 25      | 100    | null            | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0              | 0                    | the parent item is visible to the user's class |
      | 25      | 101    | null            | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0              | 0                    | the parent item is visible to the user's class |

  Scenario: Create a new groups_items row (manager access)
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following rows:
      | idGroup | idItem | bManagerAccess | bCachedManagerAccess | sAccessReason                            |
      | 21      | 100    | 1              | 1                    | the admin can manage the item's ancestor |
      | 21      | 101    | 0              | 1                    | null                                     |
      | 21      | 102    | 0              | 1                    | null                                     |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bCachedManagerAccess | sAccessReason                                  |
      | 21      | 100    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 1                    | the admin can manage the item's ancestor       |
      | 21      | 101    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 1                    | null                                           |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 1                    | null                                           |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 1                    | null                                           |
      | 23      | 102    | 2019-04-06T09:26:40Z | 2019-04-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-03-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0                    | the user really needs this access              |
      | 23      | 103    | null                 | 2019-04-06T09:26:40Z  | null                 | 2019-03-06T09:26:40Z     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | null                                           |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | the parent item is visible to the user's class |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0                    | the parent item is visible to the user's class |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | null                                           |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | null                                           |

  Scenario: Create a new groups_items row (owner access on the item's ancestor)
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following rows:
      | idGroup | idItem | bOwnerAccess | sAccessReason                                |
      | 21      | 100    | 1            | the admin is an owner of the item's ancestor |
      | 21      | 101    | 0            | null                                         |
      | 21      | 102    | 0            | null                                         |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bManagerAccess | bCachedManagerAccess | bOwnerAccess | sAccessReason                                  |
      | 21      | 100    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 0                    | 1            | the admin is an owner of the item's ancestor   |
      | 21      | 101    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 0                    | 0            | null                                           |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 0                    | 0            | null                                           |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 0                    | 0            | null                                           |
      | 23      | 102    | 2019-04-06T09:26:40Z | 2019-04-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-03-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0              | 0                    | 0            | the user really needs this access              |
      | 23      | 103    | null                 | 2019-04-06T09:26:40Z  | null                 | 2019-03-06T09:26:40Z     | null                    | null                 | 2019-05-06T09:26:40Z       | 0              | 0                    | 0            | null                                           |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0              | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0              | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 0                    | 0            | null                                           |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 0                    | 0            | null                                           |

  Scenario: Create a new groups_items row (owner access on the item)
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following row:
      | idGroup | idItem | bOwnerAccess | sAccessReason                     |
      | 21      | 102    | 1            | the admin is an owner of the item |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bCachedManagerAccess | bOwnerAccess | sAccessReason                                  |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 1            | the admin is an owner of the item              |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 23      | 102    | 2019-04-06T09:26:40Z | 2019-04-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-03-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0                    | 0            | the user really needs this access              |
      | 23      | 103    | null                 | 2019-04-06T09:26:40Z  | null                 | 2019-03-06T09:26:40Z     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |

  Scenario: Update an existing groups_items row
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following rows:
      | idGroup | idItem | sFullAccessDate | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bManagerAccess | bCachedManagerAccess | sAccessReason                                  |
      | 21      | 100    | null            | null                  | null                 | null                     | null                    | null                 | null                       | 1              | 1                    | the admin can manage the item's ancestor       |
      | 21      | 101    | null            | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 1                    | null                                           |
      | 21      | 102    | null            | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 1                    | null                                           |
      | 23      | 102    | null            | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 0                    | null                                           |
      | 23      | 103    | null            | null                  | null                 | null                     | null                    | null                 | null                       | 0              | 0                    | null                                           |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bCachedManagerAccess | sAccessReason                                  |
      | 21      | 100    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 1                    | the admin can manage the item's ancestor       |
      | 21      | 101    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 1                    | null                                           |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 1                    | null                                           |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 1                    | null                                           |
      | 23      | 102    | 2019-04-06T09:26:40Z | 2019-04-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-03-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0                    | the user really needs this access              |
      | 23      | 103    | null                 | 2019-04-06T09:26:40Z  | null                 | 2019-03-06T09:26:40Z     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | null                                           |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | the parent item is visible to the user's class |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0                    | the parent item is visible to the user's class |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | null                                           |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | null                                           |

  Scenario: Create a new groups_items row (the group has only partial access on the item's parent)
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following rows:
      | idGroup | idItem | bOwnerAccess | sPartialAccessDate   | sCachedPartialAccessDate | sAccessReason                                     |
      | 21      | 102    | 1            | null                 | null                     | the admin is an owner of the item                 |
      | 31      | 101    | 0            | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | the group has partial access to the item's parent |
    When I send a PUT request to "/groups/31/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bCachedManagerAccess | bOwnerAccess | sAccessReason                                     |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 1            | the admin is an owner of the item                 |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                              |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class    |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class    |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                              |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                              |
      | 31      | 101    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | 0            | the group has partial access to the item's parent |
      | 31      | 102    | 2019-04-06T09:26:40Z | 2019-04-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0                    | 0            | the user really needs this access                 |
      | 31      | 103    | null                 | 2019-04-06T09:26:40Z  | null                 | 2019-01-06T09:26:40Z     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                              |

  Scenario: Create a new groups_items row (the group has only full access on the item's parent)
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following rows:
      | idGroup | idItem | bOwnerAccess | sFullAccessDate      | sCachedFullAccessDate | sAccessReason                                  |
      | 21      | 102    | 1            | null                 | null                  | the admin is an owner of the item              |
      | 31      | 101    | 0            | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z  | the group has full access to the item's parent |
    When I send a PUT request to "/groups/31/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bCachedManagerAccess | bOwnerAccess | sAccessReason                                  |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 1            | the admin is an owner of the item              |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 31      | 101    | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | the group has full access to the item's parent |
      | 31      | 102    | 2019-04-06T09:26:40Z | 2019-01-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-03-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0                    | 0            | the user really needs this access              |
      | 31      | 103    | null                 | 2019-01-06T09:26:40Z  | null                 | 2019-03-06T09:26:40Z     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |


  Scenario: Create a new groups_items row (the group has no access to the item's parents, but has full access to the item itself)
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following rows:
      | idGroup | idItem | bOwnerAccess | sFullAccessDate      | sCachedFullAccessDate | sAccessReason                                  |
      | 21      | 100    | 1            | null                 | null                  | the admin is an owner of the item              |
      | 31      | 100    | 0            | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z  | the group has full access to the item's parent |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bCachedManagerAccess | bOwnerAccess | sAccessReason                                  |
      | 21      | 100    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 1            | the admin is an owner of the item              |
      | 21      | 101    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 31      | 100    | 2019-04-06T09:26:40Z | 2019-04-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-03-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0                    | 0            | the user really needs this access              |
      | 31      | 101    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | 2019-03-06T09:26:40Z    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |
      | 31      | 102    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |
      | 31      | 103    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |

  Scenario: Create a new groups_items row (the group has no access to the item's parents, but has partial access to the item itself)
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following rows:
      | idGroup | idItem | bOwnerAccess | sPartialAccessDate   | sCachedPartialAccessDate | sAccessReason                                     |
      | 21      | 100    | 1            | null                 | null                     | the admin is an owner of the item                 |
      | 31      | 100    | 0            | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | the group has partial access to the item's parent |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bCachedManagerAccess | bOwnerAccess | sAccessReason                                  |
      | 21      | 100    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 1            | the admin is an owner of the item              |
      | 21      | 101    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 31      | 100    | 2019-04-06T09:26:40Z | 2019-04-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-03-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0                    | 0            | the user really needs this access              |
      | 31      | 101    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | 2019-03-06T09:26:40Z    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |
      | 31      | 102    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |
      | 31      | 103    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |

  Scenario: Create a new groups_items row (the group has no access to the item's parents, but has grayed access to the item itself)
    Given I am the user with ID "1"
    And the database table 'groups_items' has also the following rows:
      | idGroup | idItem | bOwnerAccess | sCachedGrayedAccessDate | sAccessReason                                    |
      | 21      | 100    | 1            | null                    | the admin is an owner of the item                |
      | 31      | 100    | 0            | 2019-01-06T09:26:40Z    | the group has grayed access to the item's parent |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | idGroup | idItem | sFullAccessDate      | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bCachedManagerAccess | bOwnerAccess | sAccessReason                                  |
      | 21      | 100    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 1            | the admin is an owner of the item              |
      | 21      | 101    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 21      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 21      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 100    | null                 | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 101    | null                 | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0                    | 0            | the parent item is visible to the user's class |
      | 25      | 102    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 25      | 103    | null                 | null                  | null                 | null                     | null                    | null                 | null                       | 0                    | 0            | null                                           |
      | 31      | 100    | 2019-04-06T09:26:40Z | 2019-04-06T09:26:40Z  | 2019-03-06T09:26:40Z | 2019-03-06T09:26:40Z     | null                    | 2019-05-06T09:26:40Z | 2019-05-06T09:26:40Z       | 0                    | 0            | the user really needs this access              |
      | 31      | 101    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | 2019-03-06T09:26:40Z    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |
      | 31      | 102    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |
      | 31      | 103    | null                 | 2019-04-06T09:26:40Z  | null                 | null                     | null                    | null                 | 2019-05-06T09:26:40Z       | 0                    | 0            | null                                           |

