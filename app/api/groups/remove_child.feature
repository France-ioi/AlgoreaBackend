Feature: Remove a direct parent-child relation between two groups

  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id | owned_group_id | first_name  | last_name | allow_subgroups |
      | 1  | owner | 21            | 22             | Jean-Michel | Blanquer  | 1               |
    And the database has the following table 'groups':
      | id | name    | type      |
      | 11 | Group A | Class     |
      | 13 | Group B | Class     |
      | 14 | Group C | Class     |
      | 21 | Self    | UserSelf  |
      | 22 | Owned   | UserAdmin |

    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type   |
      | 13              | 11             | direct |
      | 22              | 11             | direct |
      | 22              | 13             | direct |
      | 22              | 14             | direct |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 11             | 0       |
      | 13                | 13             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 11             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 22             | 1       |

  Scenario: User deletes a relation
    Given I am the user with id "1"
    When I send a DELETE request to "/groups/13/relations/11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | type   | role   |
      | 22              | 11             | direct | member |
      | 22              | 13             | direct | member |
      | 22              | 14             | direct | member |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 13             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 11             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 22             | 1       |
    And the table "groups" should stay unchanged

  Scenario: User deletes a relation and an orphaned child group
    Given I am the user with id "1"
    When I send a DELETE request to "/groups/22/relations/13?delete_orphans=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | type   | role   |
      | 22              | 11             | direct | member |
      | 22              | 14             | direct | member |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 11             | 0       |
      | 22                | 14             | 0       |
      | 22                | 22             | 1       |
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" should not contain id "13"
