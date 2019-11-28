Feature: Add a parent-child relation between two groups

  Background:
    Given the database has the following table 'groups':
      | id | name    | type      |
      | 11 | Group A | Class     |
      | 13 | Group B | Class     |
      | 14 | Group C | Class     |
      | 21 | Self    | UserSelf  |
      | 22 | Owned   | UserAdmin |
    And the database has the following table 'users':
      | login | temp_user | group_id | owned_group_id | first_name  | last_name | allow_subgroups |
      | owner | 0         | 21       | 22             | Jean-Michel | Blanquer  | 1               |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 11       | 21         |
      | 13       | 21         |
      | 14       | 21         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 13             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |

  Scenario: User is an owner of the two groups and is allowed to create sub-groups
    Given I am the user with id "21"
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
      | parent_group_id | child_group_id | child_order | role   |
      | 13              | 11             | 1           | member |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 11             | 0       |
      | 13                | 13             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
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
      | parent_group_id | child_group_id | child_order | role   |
      | 13              | 11             | 1           | member |
      | 13              | 14             | 2           | member |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 11             | 0       |
      | 13                | 13             | 1       |
      | 13                | 14             | 0       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
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
      | parent_group_id | child_group_id | child_order | role   |
      | 13              | 11             | 3           | member |
      | 13              | 14             | 2           | member |
