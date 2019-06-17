Feature: Change item access rights for a group - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  |
      | 2  | user   | 23          | 24           | John        | Doe       |
      | 3  | admin  | 31          | 32           | Allie       | Grater    |
    And the database has the following table 'groups':
      | ID | sName       | sType     |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
      | 23 | user        | UserSelf  |
      | 24 | user-admin  | UserAdmin |
      | 25 | some class  | Class     |
      | 31 | admin       | UserSelf  |
      | 32 | admin-admin | UserAdmin |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
      | 22              | 23           | 0       |
      | 23              | 23           | 1       |
      | 24              | 24           | 1       |
      | 25              | 23           | 0       |
      | 25              | 25           | 1       |
      | 31              | 31           | 1       |
      | 32              | 32           | 1       |
      | 32              | 23           | 0       |
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
      | idGroup | idItem | sFullAccessDate | sCachedFullAccessDate | sPartialAccessDate   | sCachedPartialAccessDate | sCachedGrayedAccessDate | sAccessSolutionsDate | sCachedAccessSolutionsDate | bOwnerAccess | sAccessReason                                  |
      | 21      | 100    | null            | null                  | null                 | null                     | null                    | null                 | null                       | 1            | owner owns the item                            |
      | 25      | 100    | null            | null                  | 2019-01-06T09:26:40Z | 2019-01-06T09:26:40Z     | null                    | null                 | null                       | 0            | the parent item is visible to the user's class |
      | 25      | 101    | null            | null                  | null                 | null                     | 2019-01-06T09:26:40Z    | null                 | null                       | 0            | the parent item is visible to the user's class |

  Scenario: Invalid group_id
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/abc/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_items" should stay unchanged

  Scenario: Invalid item_id
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/23/items/abc" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "groups_items" should stay unchanged

  Scenario: Access reason is too long
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901"
    }
    """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "error_text": "Invalid input data",
        "errors": {
          "access_reason": ["access_reason must be a maximum of 200 characters in length"]
        },
        "message": "Bad Request",
        "success": false
      }
      """
    And the table "groups_items" should stay unchanged

  Scenario: The user is not found
    Given I am the user with ID "404"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: The user is not a manager/owner of the item
    Given I am the user with ID "3"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: The item doesn't exist
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/23/items/404" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: The user doesn't own the group
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/21/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: The group doesn't exist
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/404/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: There are no item's parents visible to the group
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/23/items/103" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

