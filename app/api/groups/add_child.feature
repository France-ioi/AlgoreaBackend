Feature: Add a parent-child relation between two groups

  Background:
    Given the database has the following table 'users':
      | id | login | temp_user | self_group_id | owned_group_id | first_name  | last_name | allow_subgroups |
      | 1  | owner | 0         | 21            | 22             | Jean-Michel | Blanquer  | 1               |
    And the database has the following table 'groups':
      | id | name    | type      |
      | 11 | Group A | Class     |
      | 13 | Group B | Class     |
      | 14 | Group C | Class     |
      | 21 | Self    | UserSelf  |
      | 22 | Owned   | UserAdmin |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type   | child_order |
      | 22              | 11             | direct | 1           |
      | 22              | 13             | direct | 1           |
      | 22              | 14             | direct | 1           |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 13             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 11             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 22             | 1       |

  Scenario: User is an owner of the two groups and is allowed to create sub-groups
    Given I am the user with id "1"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | child_order | type   | role   |
      | 13              | 11             | 1           | direct | member |
      | 22              | 11             | 1           | direct | member |
      | 22              | 13             | 1           | direct | member |
      | 22              | 14             | 1           | direct | member |
    And the table "groups_ancestors" should be:
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
    When I send a POST request to "/groups/13/relations/14"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | child_order | type   | role   |
      | 13              | 11             | 1           | direct | member |
      | 13              | 14             | 2           | direct | member |
      | 22              | 11             | 1           | direct | member |
      | 22              | 13             | 1           | direct | member |
      | 22              | 14             | 1           | direct | member |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 11             | 0       |
      | 13                | 13             | 1       |
      | 13                | 14             | 0       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 11             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 22             | 1       |
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | child_order | type   | role   |
      | 13              | 11             | 3           | direct | member |
      | 13              | 14             | 2           | direct | member |
      | 22              | 11             | 1           | direct | member |
      | 22              | 13             | 1           | direct | member |
      | 22              | 14             | 1           | direct | member |
