Feature: Change item access rights for a group - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name        | type      |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
      | 23 | user        | UserSelf  |
      | 24 | user-admin  | UserAdmin |
      | 25 | some class  | Class     |
      | 31 | admin       | UserSelf  |
      | 32 | admin-admin | UserAdmin |
    And the database has the following table 'users':
      | login | group_id | owned_group_id | first_name  | last_name |
      | owner | 21       | 22             | Jean-Michel | Blanquer  |
      | user  | 23       | 24             | John        | Doe       |
      | admin | 31       | 32             | Allie       | Grater    |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
      | 22                | 23             | 0       |
      | 23                | 23             | 1       |
      | 24                | 24             | 1       |
      | 25                | 23             | 0       |
      | 25                | 25             | 1       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 32                | 23             | 0       |
    And the database has the following table 'items':
      | id  |
      | 100 |
      | 101 |
      | 102 |
      | 103 |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | partial_access_propagation | child_order |
      | 100            | 101           | AsGrayed                   | 0           |
      | 101            | 102           | AsPartial                  | 0           |
      | 102            | 103           | AsPartial                  | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 100              | 101           |
      | 100              | 102           |
      | 100              | 103           |
      | 101              | 102           |
      | 101              | 103           |
      | 102              | 103           |
    And the database has the following table 'groups_items':
      | group_id | item_id | full_access_since | cached_full_access_since | partial_access_since | cached_partial_access_since | cached_grayed_access_since | solutions_access_since | cached_solutions_access_since | owner_access | access_reason                                  | creator_user_group_id |
      | 21       | 100     | null              | null                     | null                 | null                        | null                       | null                   | null                          | 1            | owner owns the item                            | 23                    |
      | 21       | 101     | null              | null                     | null                 | null                        | null                       | null                   | null                          | 0            | null                                           | 23                    |
      | 21       | 102     | null              | null                     | null                 | null                        | null                       | null                   | null                          | 0            | null                                           | 23                    |
      | 21       | 103     | null              | null                     | null                 | null                        | null                       | null                   | null                          | 0            | null                                           | 23                    |
      | 25       | 100     | null              | null                     | 2019-01-06 09:26:40  | 2019-01-06 09:26:40         | null                       | null                   | null                          | 0            | the parent item is visible to the user's class | 23                    |
      | 25       | 101     | null              | null                     | null                 | null                        | 2019-01-06 09:26:40        | null                   | null                          | 0            | null                                           | 23                    |

  Scenario: Invalid group_id
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/abc/items/102" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_items" should stay unchanged

  Scenario: Invalid item_id
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/23/items/abc" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "groups_items" should stay unchanged

  Scenario: Access reason is too long
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
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

  Scenario: The user doesn't exist
    Given I am the user with group_id "404"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups_items" should stay unchanged

  Scenario: The user is not a manager/owner of the item
    Given I am the user with group_id "31"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: The item doesn't exist
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/23/items/404" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: The user doesn't own the group
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/21/items/102" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: The group doesn't exist
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/404/items/102" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged

  Scenario: There are no item's parents visible to the group
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/23/items/103" with the following body:
    """
    {
      "partial_access_since": "2019-03-06T09:26:40Z",
      "full_access_since": "2019-04-06T09:26:40Z",
      "solutions_access_since": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_items" should stay unchanged
