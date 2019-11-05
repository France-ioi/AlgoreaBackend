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
      | parent_item_id | child_item_id | content_view_propagation | child_order |
      | 100            | 101           | as_info                  | 0           |
      | 101            | 102           | as_content               | 0           |
      | 102            | 103           | as_content               | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 100              | 101           |
      | 100              | 102           |
      | 100              | 103           |
      | 101              | 102           |
      | 101              | 103           |
      | 102              | 103           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_grant_view_generated | is_owner_generated |
      | 21       | 100     | solution           | transfer                 | 1                  |
      | 21       | 101     | none               | none                     | 0                  |
      | 21       | 102     | none               | solution                 | 0                  |
      | 21       | 103     | none               | solution                 | 0                  |
      | 25       | 100     | content            | none                     | 0                  |
      | 25       | 101     | info               | none                     | 0                  |
      | 31       | 102     | none               | content_with_descendants | 0                  |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | can_grant_view           | is_owner | giver_group_id |
      | 21       | 100     | none     | none                     | 1        | 23             |
      | 21       | 102     | none     | solution                 | 1        | 23             |
      | 25       | 100     | content  | none                     | 0        | 23             |
      | 25       | 101     | info     | none                     | 0        | 23             |
      | 31       | 102     | none     | content_with_descendants | 0        | 31             |

  Scenario: Invalid group_id
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/abc/items/102" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Invalid item_id
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/23/items/abc" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: Invalid can_view
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "can_view": "unknown"
    }
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "error_text": "Invalid input data",
      "errors": {
        "can_view": ["can_view must be one of [none info content content_with_descendants solution]"]
      },
      "message": "Bad Request",
      "success": false
    }
    """
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: The user doesn't exist
    Given I am the user with group_id "404"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: The user doesn't have enough rights to set can_view
    Given I am the user with group_id "31"
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: The user doesn't have enough rights to set can_view = info
    Given I am the user with group_id "31"
    When I send a PUT request to "/groups/23/items/103" with the following body:
    """
    {
      "can_view": "info"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: The item doesn't exist
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/23/items/404" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: The user doesn't own the group
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/21/items/102" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: The group doesn't exist
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/404/items/102" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged

  Scenario: There are no item's parents visible to the group
    Given I am the user with group_id "21"
    When I send a PUT request to "/groups/23/items/103" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should stay unchanged
